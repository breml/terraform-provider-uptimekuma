package client

import (
	"context"
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

// TestAuthSchemeProbeError_RealClient pins the classification of the error
// the real client reports when its auth scheme detection gets an answer no
// Uptime Kuma gives, such as an access proxy refusing the request. It is
// retried, because a proxy's refusal may be passing.
func TestAuthSchemeProbeError_RealClient(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusForbidden)
	}))
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

	if attempts != 2 {
		t.Errorf("expected the probe error to be retried, got %d attempt(s)", attempts)
	}
}
