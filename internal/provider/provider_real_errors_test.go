package provider

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// redirectTo returns a server that redirects every request to the URL target
// builds from the server's own URL.
func redirectTo(t *testing.T, target func(serverURL string) string) string {
	t.Helper()

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target(server.URL)+r.URL.Path, http.StatusFound)
	}))
	t.Cleanup(server.Close)

	return server.URL
}

// TestConnectionError_RealClient pins the diagnostic for the errors the real
// client reports for a request configuration it rejects, a redirect it refuses
// and an auth scheme detection answered by something other than Uptime Kuma.
// The client has no sentinels for them, so this is what keeps the advice in
// step with its messages.
func TestConnectionError_RealClient(t *testing.T) {
	t.Parallel()

	forbidden := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(forbidden.Close)

	otherHost := redirectTo(t, func(serverURL string) string {
		return strings.Replace(serverURL, "127.0.0.1", "localhost", 1)
	})
	httpsUpgrade := redirectTo(t, func(serverURL string) string {
		return strings.Replace(serverURL, "http://", "https://", 1)
	})

	tests := []struct {
		name        string
		config      client.Config
		wantSummary string
		wantDetail  string
		notDetail   string
	}{
		{
			name:        "reserved header",
			config:      client.Config{Headers: map[string]string{"Host": "example.com"}},
			wantSummary: "Invalid api_key, headers or user_agent",
		},
		{
			name:        "authorization bearer header",
			config:      client.Config{Headers: map[string]string{"Authorization": "Bearer abc"}},
			wantSummary: "Invalid api_key, headers or user_agent",
		},
		{
			name:        "bearer prefixed key",
			config:      client.Config{APIKey: "Bearer uk1_abc"},
			wantSummary: "Invalid api_key, headers or user_agent",
		},
		{
			name:        "key with a space",
			config:      client.Config{APIKey: "uk1_abc def"},
			wantSummary: "Invalid api_key, headers or user_agent",
			wantDetail:  "trimspace()",
		},
		{
			name:        "userinfo next to a key",
			config:      client.Config{Endpoint: "http://user:p4ssw0rd@127.0.0.1:1", APIKey: "uk1_abc"},
			wantSummary: "endpoint contains a user name and password",
			wantDetail:  "`endpoint`",
			notDetail:   "p4ssw0rd",
		},
		{
			name:        "invalid user agent",
			config:      client.Config{UserAgent: "bad\nagent"},
			wantSummary: "Invalid api_key, headers or user_agent",
		},
		{
			name:        "redirect to another host",
			config:      client.Config{Endpoint: otherHost},
			wantSummary: "Uptime Kuma redirected to another origin",
			notDetail:   "service token",
		},
		{
			name:        "redirect to another host with headers",
			config:      client.Config{Endpoint: otherHost, Headers: map[string]string{"X-Token": "t"}},
			wantSummary: "Uptime Kuma redirected to another origin",
			wantDetail:  "service token in `headers`",
		},
		{
			name:        "redirect to https",
			config:      client.Config{Endpoint: httpsUpgrade},
			wantSummary: "Uptime Kuma redirected to https",
		},
		{
			name:        "auth scheme detection refused by a proxy",
			config:      client.Config{Endpoint: forbidden.URL},
			wantSummary: "Unexpected answer to the auth scheme detection",
			wantDetail:  "`auth_scheme`",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := tc.config
			if config.Endpoint == "" {
				config.Endpoint = "http://127.0.0.1:1"
			}

			config.ConnectTimeout = 10 * time.Second

			_, err := client.New(t.Context(), &config)
			if err == nil {
				t.Fatal("expected the connection to fail")
			}

			summary, detail := connectionError(config.Endpoint, err, requestSecretsOf(&config))
			if summary != tc.wantSummary {
				t.Errorf("expected summary %q, got %q: %s", tc.wantSummary, summary, detail)
			}

			if !strings.Contains(detail, tc.wantDetail) {
				t.Errorf("expected detail to contain %q, got:\n%s", tc.wantDetail, detail)
			}

			if tc.notDetail != "" && strings.Contains(detail, tc.notDetail) {
				t.Errorf("expected detail not to contain %q, got:\n%s", tc.notDetail, detail)
			}
		})
	}
}

// TestConfigure_SendsUserAgentAndHeaders pins that the first request
// Configure makes carries the User-Agent, by default the provider's own with
// its version, and the configured headers. The server refuses the request, so
// the connection fails afterwards; only the request matters.
func TestConfigure_SendsUserAgentAndHeaders(t *testing.T) {
	str := func(value string) tftypes.Value { return tftypes.NewValue(tftypes.String, value) }

	tests := []struct {
		name          string
		values        map[string]tftypes.Value
		wantUserAgent string
		wantHeader    string
	}{
		{name: "default user agent", wantUserAgent: "terraform-provider-uptimekuma/1.2.3"},
		{
			name:          "configured user agent",
			values:        map[string]tftypes.Value{"user_agent": str("my-pipeline/1.0")},
			wantUserAgent: "my-pipeline/1.0",
		},
		{
			name: "configured header",
			values: map[string]tftypes.Value{
				"headers": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
					"X-Acceptance": str("seen"),
				}),
			},
			wantUserAgent: "terraform-provider-uptimekuma/1.2.3",
			wantHeader:    "seen",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearCredentialEnvironment(t)
			t.Setenv("UPTIMEKUMA_ENDPOINT", "")
			client.ResetGlobalPool()
			t.Cleanup(client.ResetGlobalPool)

			var (
				once  sync.Once
				first http.Header
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				once.Do(func() { first = r.Header.Clone() })
				w.WriteHeader(http.StatusForbidden)
			}))
			t.Cleanup(server.Close)

			values := map[string]tftypes.Value{
				"endpoint":    str(server.URL),
				"timeout":     str("5s"),
				"max_retries": tftypes.NewValue(tftypes.Number, 0),
			}
			maps.Copy(values, tc.values)

			req := provider.ConfigureRequest{Config: providerConfigValue(t, values)}

			var resp provider.ConfigureResponse

			(&UptimeKumaProvider{version: "1.2.3"}).Configure(t.Context(), req, &resp)

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected the refused connection to fail Configure")
			}

			if first == nil {
				t.Fatal("expected the server to receive a request")
			}

			if got := first.Get("User-Agent"); got != tc.wantUserAgent {
				t.Errorf("expected User-Agent %q, got %q", tc.wantUserAgent, got)
			}

			if got := first.Get("X-Acceptance"); got != tc.wantHeader {
				t.Errorf("expected X-Acceptance %q, got %q", tc.wantHeader, got)
			}
		})
	}
}
