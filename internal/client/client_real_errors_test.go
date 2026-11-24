package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// countingConnect returns kuma.New as a connectFunc that counts its calls, so
// a test can tell an error the retry loop gives up on at once from one it
// retries.
func countingConnect(attempts *int) connectFunc {
	return func(
		ctx context.Context, baseURL string, username string, password string, opts ...kuma.Option,
	) (*kuma.Client, error) {
		*attempts++

		return kuma.New(ctx, baseURL, username, password, opts...)
	}
}

// redirectingServer returns a server that answers every request with a
// redirect to the URL target builds from the server's own URL.
func redirectingServer(t *testing.T, target func(serverURL string) string) string {
	t.Helper()

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target(server.URL)+r.URL.Path, http.StatusFound)
	}))
	t.Cleanup(server.Close)

	return server.URL
}

// TestRequestConfigErrors_RealClient pins the classification of the errors
// the real kuma.New rejects a request configuration with, which the client
// reports without a sentinel. Each must be recognised by IsRequestConfigError
// and end the retries after the first attempt. None of them reaches a server.
func TestRequestConfigErrors_RealClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		config       Config
		wantUserinfo bool
	}{
		{name: "reserved header", config: Config{Headers: map[string]string{"Host": "example.com"}}},
		{name: "authorization bearer header", config: Config{Headers: map[string]string{"Authorization": "Bearer abc"}}},
		{name: "bearer prefixed key", config: Config{APIKey: "Bearer uk1_abc"}},
		{name: "key with a space", config: Config{APIKey: "uk1_abc def"}},
		{name: "key with a trailing newline", config: Config{APIKey: "uk1_abc\n"}},
		{
			name:         "userinfo next to a key",
			config:       Config{Endpoint: "http://user:pass@127.0.0.1:1", APIKey: "uk1_abc"},
			wantUserinfo: true,
		},
		{name: "invalid user agent", config: Config{UserAgent: "bad\nagent"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := tc.config
			if config.Endpoint == "" {
				config.Endpoint = "http://127.0.0.1:1"
			}

			config.ConnectTimeout = 10 * time.Second
			config.MaxRetries = 3

			var attempts int

			_, err := newClientDirectWithRetry(t.Context(), &config, countingConnect(&attempts), noDelay)
			if err == nil {
				t.Fatal("expected kuma.New to reject the configuration")
			}

			if !IsRequestConfigError(err) {
				t.Errorf("expected IsRequestConfigError for %q", err)
			}

			if got := IsUserinfoError(err); got != tc.wantUserinfo {
				t.Errorf("expected IsUserinfoError %t for %q", tc.wantUserinfo, err)
			}

			if attempts != 1 {
				t.Errorf("expected exactly one attempt, got %d", attempts)
			}
		})
	}
}

// TestRefusedRedirect_RealClient pins the classification of the redirects
// the real client refuses: to another host, and from http to https on the
// same host. Both end the retries after the first attempt.
func TestRefusedRedirect_RealClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		target    func(serverURL string) string
		wantHTTPS bool
	}{
		{
			name: "another host",
			target: func(serverURL string) string {
				return strings.Replace(serverURL, "127.0.0.1", "localhost", 1)
			},
		},
		{
			name: "http to https on the same host",
			target: func(serverURL string) string {
				return strings.Replace(serverURL, "http://", "https://", 1)
			},
			wantHTTPS: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := Config{
				Endpoint:       redirectingServer(t, tc.target),
				ConnectTimeout: 10 * time.Second,
				MaxRetries:     3,
			}

			var attempts int

			_, err := newClientDirectWithRetry(t.Context(), &config, countingConnect(&attempts), noDelay)

			if !IsRefusedRedirect(err) {
				t.Fatalf("expected IsRefusedRedirect for %v", err)
			}

			if got := IsHTTPSUpgradeRedirect(err); got != tc.wantHTTPS {
				t.Errorf("expected IsHTTPSUpgradeRedirect %t for %q", tc.wantHTTPS, err)
			}

			if IsRequestConfigError(err) || IsAuthSchemeProbeError(err) {
				t.Errorf("expected only the redirect classification for %q", err)
			}

			if attempts != 1 {
				t.Errorf("expected exactly one attempt, got %d", attempts)
			}
		})
	}
}

// TestAuthSchemeProbeError_RealClient pins the classification of the errors
// the real client reports when its auth scheme detection gets an answer no
// Uptime Kuma gives. A 4xx is a proxy refusing the request and ends the
// retries; a gateway's 408, 425 or 429, a 5xx and a redirect within the origin
// may pass and are retried.
func TestAuthSchemeProbeError_RealClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		wantAttempts int
	}{
		{
			name: "403 from an access proxy",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusForbidden)
			},
			wantAttempts: 1,
		},
		{
			// Without better-auth's X-Retry-After, a JSON 429 is not Uptime Kuma.
			name: "JSON 429 from a rate-limiting gateway",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
			},
			wantAttempts: 2,
		},
		{
			name: "502 from a reverse proxy",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadGateway)
			},
			wantAttempts: 2,
		},
		{
			name: "redirect within the origin",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth/ok" {
					http.Redirect(w, r, "/sign-in", http.StatusFound)

					return
				}

				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte("<html></html>"))
			},
			wantAttempts: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tc.handler)
			t.Cleanup(server.Close)

			config := Config{Endpoint: server.URL, ConnectTimeout: 10 * time.Second, MaxRetries: 1}

			var attempts int

			_, err := newClientDirectWithRetry(t.Context(), &config, countingConnect(&attempts), noDelay)

			if !IsAuthSchemeProbeError(err) {
				t.Fatalf("expected IsAuthSchemeProbeError for %v", err)
			}

			if IsRequestConfigError(err) || IsRefusedRedirect(err) {
				t.Errorf("expected only the probe classification for %q", err)
			}

			if attempts != tc.wantAttempts {
				t.Errorf("expected %d attempt(s), got %d", tc.wantAttempts, attempts)
			}
		})
	}
}

// TestRequestErrorMatchers pins the matchers on the client's message texts,
// copied from go-uptime-kuma-client v0.7.6.
func TestRequestErrorMatchers(t *testing.T) {
	t.Parallel()

	const (
		userinfo = "base url: userinfo (user:password@) sets an Authorization header of its own, which cannot " +
			"be combined with WithAPIKey or an Authorization header from WithHeaders"
		httpsUpgrade = `Get "http://kuma/api/auth/ok": redirect from http://kuma/api/auth/ok to ` +
			"https://kuma/api/auth/ok upgrades to https; use https://kuma as the base URL"
		crossOrigin = `Get "http://a/api/auth/ok": redirect from http://a/api/auth/ok to http://b/login ` +
			"leaves the server's origin; connect to the server on http://b"
		probeRedirect = "connect to server: detect auth scheme: GET http://kuma/api/auth/ok was redirected to " +
			"http://kuma/sign-in, which neither Uptime Kuma 2.x nor 3.x does; an access proxy in front of the " +
			"server may be redirecting to its sign-in (check WithHeaders), or pin the scheme with WithAuthScheme"
		probeStatus = "connect to server: detect auth scheme: GET http://kuma/api/auth/ok answered status %d " +
			`with Content-Type "text/html", which is neither Uptime Kuma 2.x nor 3.x; an access proxy in front ` +
			"of the server may be rejecting the request (check WithHeaders), or pin the scheme with WithAuthScheme"
	)

	tests := []struct {
		name                                             string
		err                                              error
		userinfo, https, redirect, probe, probePermanent bool
	}{
		{name: "nil"},
		{name: "unrelated", err: errors.New("dial tcp: connection refused")},
		{name: "userinfo", err: errors.New(userinfo), userinfo: true},
		{name: "userinfo wrapped", err: fmt.Errorf("login failed: %w", errors.New(userinfo)), userinfo: true},
		{name: "https upgrade", err: errors.New(httpsUpgrade), https: true, redirect: true},
		{name: "cross origin", err: errors.New(crossOrigin), redirect: true},
		{name: "probe redirect", err: errors.New(probeRedirect), probe: true},
		{name: "probe 401", err: fmt.Errorf(probeStatus, 401), probe: true, probePermanent: true},
		{name: "probe 403", err: fmt.Errorf(probeStatus, 403), probe: true, probePermanent: true},
		{name: "probe 404", err: fmt.Errorf(probeStatus, 404), probe: true, probePermanent: true},
		// A gateway's "try again later" answers may pass.
		{name: "probe 408", err: fmt.Errorf(probeStatus, 408), probe: true},
		{name: "probe 425", err: fmt.Errorf(probeStatus, 425), probe: true},
		{name: "probe 429", err: fmt.Errorf(probeStatus, 429), probe: true},
		{name: "probe 500", err: fmt.Errorf(probeStatus, 500), probe: true},
		{name: "probe 503", err: fmt.Errorf(probeStatus, 503), probe: true},
		{name: "probe transport error", err: errors.New("connect to server: detect auth scheme: dial tcp: refused")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for name, got := range map[string][2]bool{
				"IsUserinfoError":        {IsUserinfoError(tc.err), tc.userinfo},
				"IsHTTPSUpgradeRedirect": {IsHTTPSUpgradeRedirect(tc.err), tc.https},
				"IsRefusedRedirect":      {IsRefusedRedirect(tc.err), tc.redirect},
				"IsAuthSchemeProbeError": {IsAuthSchemeProbeError(tc.err), tc.probe},
				"IsPermanentProbeError":  {IsPermanentProbeError(tc.err), tc.probePermanent},
			} {
				if got[0] != got[1] {
					t.Errorf("%s(%v): expected %t, got %t", name, tc.err, got[1], got[0])
				}
			}
		})
	}
}
