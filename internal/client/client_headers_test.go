package client

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// TestNewClientDirectWithRetry_RequestHeaderOptions pins that the API key, the
// headers and the User-Agent of the Config reach kuma.New, each only once it is
// configured.
func TestNewClientDirectWithRetry_RequestHeaderOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        Config
		wantAPIKey    string
		wantHeaders   http.Header
		wantUserAgent string
	}{
		{name: "none"},
		{name: "api key", config: Config{APIKey: "uk1_key"}, wantAPIKey: "uk1_key"},
		{
			name:        "headers",
			config:      Config{Headers: map[string]string{"cf-access-client-id": "id", "X-Other": "x"}},
			wantHeaders: http.Header{"Cf-Access-Client-Id": {"id"}, "X-Other": {"x"}},
		},
		{name: "user agent", config: Config{UserAgent: "ua/1"}, wantUserAgent: "ua/1"},
		{
			// Two spellings of one name keep both values, in the sorted order of
			// the spellings.
			name:        "two spellings of one name",
			config:      Config{Headers: map[string]string{"x-a": "lower", "X-A": "upper"}},
			wantHeaders: http.Header{"X-A": {"upper", "lower"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := tc.config
			config.Endpoint = "http://uptime-kuma.invalid"
			config.ConnectTimeout = 30 * time.Second

			var captured []kuma.Option

			_, err := newClientDirectWithRetry(t.Context(), &config, capturingConnect(&captured), retryDelay)
			if !errors.Is(err, kuma.ErrInvalidCredentials) {
				t.Fatalf("expected the fake's error, got %v", err)
			}

			client := appliedOptions(captured)

			if got := field(t, client, "apiKey").String(); got != tc.wantAPIKey {
				t.Errorf("expected api key %q, got %q", tc.wantAPIKey, got)
			}

			if got := field(t, client, "userAgent").String(); got != tc.wantUserAgent {
				t.Errorf("expected user agent %q, got %q", tc.wantUserAgent, got)
			}

			got := headerField(field(t, client, "headers"))
			if len(got) != len(tc.wantHeaders) || (len(got) > 0 && !reflect.DeepEqual(got, tc.wantHeaders)) {
				t.Errorf("expected headers %v, got %v", tc.wantHeaders, got)
			}
		})
	}
}

// headerField copies the unexported http.Header field of the client, which
// reflection cannot hand out as an interface.
func headerField(value reflect.Value) http.Header {
	header := http.Header{}

	for _, name := range value.MapKeys() {
		values := value.MapIndex(name)
		for i := range values.Len() {
			header[name.String()] = append(header[name.String()], values.Index(i).String())
		}
	}

	return header
}

// TestPool_ConfigMatches_RequestHeaders pins the API key, the headers and the
// User-Agent as part of the pool's connection identity: a provider block that
// authenticates by an API key must never be handed the connection a block with
// a username and password opened, and the other way round.
func TestPool_ConfigMatches_RequestHeaders(t *testing.T) {
	t.Parallel()

	base := Config{
		Endpoint:  "http://localhost:3001",
		APIKey:    "uk1_key",
		Headers:   map[string]string{"CF-Access-Client-Id": "id", "CF-Access-Client-Secret": "secret"},
		UserAgent: "terraform-provider-uptimekuma/test",
	}

	pool := &Pool{config: &base}

	same := base

	if !pool.configMatches(&same) {
		t.Error("expected configMatches to return true for identical headers")
	}

	// The provider hands over nil for no headers, but an empty map means the
	// same, and must not force a second connection.
	withoutHeaders := &Pool{config: &Config{Endpoint: base.Endpoint, Headers: nil}}
	if !withoutHeaders.configMatches(&Config{Endpoint: base.Endpoint, Headers: map[string]string{}}) {
		t.Error("expected configMatches to treat nil and empty headers alike")
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "other api key", mutate: func(c *Config) { c.APIKey = "uk1_other" }},
		{
			name: "username and password instead of api key",
			mutate: func(c *Config) {
				c.APIKey = ""
				c.Username = "admin"
				c.Password = "secret"
			},
		},
		{
			name:   "other header value",
			mutate: func(c *Config) { c.Headers = map[string]string{"CF-Access-Client-Id": "id"} },
		},
		{name: "no headers", mutate: func(c *Config) { c.Headers = nil }},
		{name: "other user agent", mutate: func(c *Config) { c.UserAgent = "other/1" }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requested := base
			tc.mutate(&requested)

			if pool.configMatches(&requested) {
				t.Error("expected configMatches to return false")
			}
		})
	}
}

// TestPool_GetOrCreate_MismatchRedactsAPIKeyAndHeaders pins that the mismatch
// message names the API key and the headers without their values: a header
// such as CF-Access-Client-Secret is a credential.
func TestPool_GetOrCreate_MismatchRedactsAPIKeyAndHeaders(t *testing.T) {
	t.Parallel()

	const (
		apiKey       = "uk1_secret-api-key"
		headerSecret = "header-secret-value"
	)

	pool := &Pool{
		client: &kuma.Client{},
		config: &Config{
			Endpoint: "http://localhost:3001",
			Username: "admin",
			Password: "secret",
		},
	}

	_, err := pool.GetOrCreate(t.Context(), &Config{
		Endpoint:  "http://localhost:3001",
		APIKey:    apiKey,
		Headers:   map[string]string{"CF-Access-Client-Secret": headerSecret, "CF-Access-Client-Id": "id"},
		UserAgent: "ua/1",
	})
	if err == nil {
		t.Fatal("expected a config mismatch error, got nil")
	}

	message := err.Error()

	for _, secret := range []string{apiKey, headerSecret, "secret\""} {
		if strings.Contains(message, secret) {
			t.Errorf("expected the mismatch message not to carry %q, got %q", secret, message)
		}
	}

	for _, want := range []string{
		`api_key=unset headers=[] user_agent=""`,
		`api_key=set headers=[CF-Access-Client-Id CF-Access-Client-Secret] user_agent="ua/1"`,
		`differing secrets: password, api_key, headers`,
	} {
		if !strings.Contains(message, want) {
			t.Errorf("expected the mismatch message to contain %q, got %q", want, message)
		}
	}
}

// TestIsPermanentLoginError_APIKeyAndRequestConfig pins that what an API key
// login or a rejected request configuration fails with ends the retries: no
// further attempt with the same configuration can succeed.
func TestIsPermanentLoginError_APIKeyAndRequestConfig(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		fmt.Errorf("connect: %w", kuma.ErrAPIKeyUnsupported),
		fmt.Errorf("connect: %w", kuma.ErrAPIKeyRejected),
		fmt.Errorf("connect: %w", kuma.ErrAPIKeySetupRequired),
		errors.New("headers: Host is set by the client or its transport and cannot be configured"),
		errors.New("api key: pass the key without the Bearer prefix"),
		errors.New("base url: userinfo (user:password@) sets an Authorization header of its own"),
		errors.New(`Get "http://kuma/api/entry-page": redirect from http://kuma/api/entry-page to ` +
			"https://kuma/api/entry-page upgrades to https; use https://kuma as the base URL"),
		errors.New(`Get "http://a/x": redirect from http://a/x to http://b/x leaves the server's origin; ` +
			"connect to the server on http://b"),
	} {
		if !isPermanentLoginError(err) {
			t.Errorf("expected %q to be permanent", err)
		}
	}

	for _, err := range []error{
		errors.New("dial tcp: connection refused"),
		errors.New("wait for ready events: timeout"),
	} {
		if isPermanentLoginError(err) {
			t.Errorf("expected %q to be retried", err)
		}
	}
}

func TestIsRequestConfigError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  error
		want bool
	}{
		{err: nil},
		{err: errors.New("headers: X: not a valid header value"), want: true},
		{err: fmt.Errorf("login failed: %w", errors.New("headers: X: not a valid header value")), want: true},
		{err: errors.New("api key: the key contains whitespace or control characters"), want: true},
		{err: errors.New("headers: WithUserAgent: not a valid header value"), want: true},
		// The server's answers share the "api key: " prefix, but are not a
		// rejected configuration.
		{err: fmt.Errorf("login failed: %w", kuma.ErrAPIKeyUnsupported)},
		{err: fmt.Errorf("login failed: %w", kuma.ErrAPIKeySetupRequired)},
		{err: errors.New("dial tcp: connection refused")},
	}

	for _, tc := range tests {
		if got := IsRequestConfigError(tc.err); got != tc.want {
			t.Errorf("IsRequestConfigError(%v): expected %t, got %t", tc.err, tc.want, got)
		}
	}
}

// TestPool_GetOrCreate_MismatchNamesDifferingHeaderValue pins that a header
// whose value alone differs is named among the differing secrets, since the
// two halves of the message print the same header names.
func TestPool_GetOrCreate_MismatchNamesDifferingHeaderValue(t *testing.T) {
	t.Parallel()

	pool := &Pool{
		client: &kuma.Client{},
		config: &Config{
			Endpoint: "http://localhost:3001",
			Headers:  map[string]string{"CF-Access-Client-Secret": "old-secret"},
		},
	}

	_, err := pool.GetOrCreate(t.Context(), &Config{
		Endpoint: "http://localhost:3001",
		Headers:  map[string]string{"CF-Access-Client-Secret": "new-secret"},
	})
	if err == nil {
		t.Fatal("expected a config mismatch error, got nil")
	}

	message := err.Error()

	if !strings.HasSuffix(message, "differing secrets: headers") {
		t.Errorf("expected the headers named as differing, got %q", message)
	}

	if strings.Contains(message, "old-secret") || strings.Contains(message, "new-secret") {
		t.Errorf("expected no header value in the message, got %q", message)
	}
}

// TestPool_GetOrCreate_MismatchRedactsEndpoint pins that the password of an
// endpoint with userinfo stays out of the mismatch message.
func TestPool_GetOrCreate_MismatchRedactsEndpoint(t *testing.T) {
	t.Parallel()

	pool := &Pool{client: &kuma.Client{}, config: &Config{Endpoint: "http://admin:s3cr3t@localhost:3001"}}

	_, err := pool.GetOrCreate(t.Context(), &Config{Endpoint: "http://admin:other@localhost:3001"})
	if err == nil {
		t.Fatal("expected a config mismatch error, got nil")
	}

	if message := err.Error(); strings.Contains(message, "s3cr3t") || strings.Contains(message, "other@") {
		t.Errorf("expected the endpoint passwords redacted, got %q", message)
	}
}
