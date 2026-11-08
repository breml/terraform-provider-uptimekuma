package client

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// appliedOptions applies opts to a zero kuma.Client and returns it as a
// reflect.Value, so a test can read what the options configured.
//
// kuma.Client exposes none of the fields its options set, so they are read by
// reflection. That ties this test to the client's internals: a field renamed in
// a client release fails the test loudly through field, rather than letting an
// option go unchecked.
func appliedOptions(opts []kuma.Option) reflect.Value {
	var c kuma.Client

	for _, opt := range opts {
		opt(&c)
	}

	return reflect.ValueOf(&c).Elem()
}

// field returns the unexported field name of the client, failing the test if
// it does not exist.
func field(t *testing.T, client reflect.Value, name string) reflect.Value {
	t.Helper()

	value := client.FieldByName(name)
	if !value.IsValid() {
		t.Fatalf("kuma.Client has no field %q, has the client been upgraded?", name)
	}

	return value
}

// capturingConnect returns a connectFunc that records the options of the
// attempt it is called for and fails it with a permanent login error, so the
// retry loop makes exactly one attempt.
func capturingConnect(captured *[]kuma.Option) connectFunc {
	return func(_ context.Context, _ string, _ string, _ string, opts ...kuma.Option) (*kuma.Client, error) {
		*captured = opts

		return nil, kuma.ErrInvalidCredentials
	}
}

// TestNewClientDirectWithRetry_CredentialOptions pins that the credentials of
// the Config reach kuma.New as the options that carry them, inspected on the
// options the connectFunc receives, and that each one is only passed once its
// own field is configured. A client without them keeps the plain password
// login, and kuma.New is never handed the empty TOTP secret it rejects
// outright.
func TestNewClientDirectWithRetry_CredentialOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		config           Config
		wantSessionToken string
		wantTOTP         bool
		wantAuthScheme   kuma.AuthScheme
		wantCallback     bool
	}{
		{name: "no credentials"},
		{
			name:     "totp secret",
			config:   Config{TOTPSecret: "JBSWY3DPEHPK3PXP"},
			wantTOTP: true,
		},
		{
			name:             "session token",
			config:           Config{SessionToken: "token"},
			wantSessionToken: "token",
		},
		{
			name:           "auth scheme http",
			config:         Config{AuthScheme: kuma.AuthSchemeHTTP},
			wantAuthScheme: kuma.AuthSchemeHTTP,
		},
		{
			name:           "auth scheme socket",
			config:         Config{AuthScheme: kuma.AuthSchemeSocket},
			wantAuthScheme: kuma.AuthSchemeSocket,
		},
		{
			name:         "rejection callback",
			config:       Config{OnSessionTokenRejected: func(error) {}},
			wantCallback: true,
		},
		{
			name: "all of them",
			config: Config{
				TOTPSecret:             "JBSWY3DPEHPK3PXP",
				SessionToken:           "token",
				AuthScheme:             kuma.AuthSchemeHTTP,
				OnSessionTokenRejected: func(error) {},
			},
			wantSessionToken: "token",
			wantTOTP:         true,
			wantAuthScheme:   kuma.AuthSchemeHTTP,
			wantCallback:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := tc.config
			config.Endpoint = "http://uptime-kuma.invalid"
			config.Username = "admin"
			config.Password = "secret"
			config.ConnectTimeout = 30 * time.Second

			var captured []kuma.Option

			_, err := newClientDirectWithRetry(t.Context(), &config, capturingConnect(&captured))
			if !errors.Is(err, kuma.ErrInvalidCredentials) {
				t.Fatalf("expected the fake's error, got %v", err)
			}

			client := appliedOptions(captured)

			if got := field(t, client, "sessionTokenPreset").String(); got != tc.wantSessionToken {
				t.Errorf("expected session token %q, got %q", tc.wantSessionToken, got)
			}

			if got := !field(t, client, "totpCode").IsNil(); got != tc.wantTOTP {
				t.Errorf("expected a totp code source %t, got %t", tc.wantTOTP, got)
			}

			wantSources := int64(0)
			if tc.wantTOTP {
				wantSources = 1
			}

			if got := field(t, client, "totpSources").Int(); got != wantSources {
				t.Errorf("expected %d totp source(s), got %d", wantSources, got)
			}

			if got := kuma.AuthScheme(field(t, client, "authScheme").Int()); got != tc.wantAuthScheme {
				t.Errorf("expected auth scheme %s, got %s", tc.wantAuthScheme, got)
			}

			if got := !field(t, client, "sessionTokenRejectedCallback").IsNil(); got != tc.wantCallback {
				t.Errorf("expected a rejection callback %t, got %t", tc.wantCallback, got)
			}
		})
	}
}

// TestNew_MalformedTOTPSecretIsNotRetried pins that a secret no code can be
// derived from fails before the first connection attempt. Retrying it would
// spend the server's login budget on a value that cannot improve, and report
// the result as a connection failure.
func TestNew_MalformedTOTPSecretIsNotRetried(t *testing.T) {
	t.Parallel()

	start := time.Now()

	_, err := New(t.Context(), &Config{Endpoint: "http://127.0.0.1:1", TOTPSecret: "JBSWY3DPEHPK3PX0"})
	if !errors.Is(err, kuma.ErrInvalidTOTPSecret) {
		t.Fatalf("expected kuma.ErrInvalidTOTPSecret, got %v", err)
	}

	// The first backoff alone is 500ms, so anything near it means the secret
	// went down the retry path.
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("expected the secret to be rejected before connecting, took %s", elapsed)
	}
}

// TestNew_MalformedTOTPSecretPooledLeavesNoPool pins that the TOTP secret
// check runs ahead of the pool too, which the provider always uses: the
// secret fails before a connection is attempted, and the pool is left without
// a client or a config a later, valid configuration would be compared with.
func TestNew_MalformedTOTPSecretPooledLeavesNoPool(t *testing.T) {
	ResetGlobalPool()
	t.Cleanup(ResetGlobalPool)

	_, err := New(t.Context(), &Config{
		Endpoint:             "http://127.0.0.1:1",
		Username:             "admin",
		Password:             "secret",
		TOTPSecret:           "JBSWY3DPEHPK3PX0",
		EnableConnectionPool: true,
	})
	if !errors.Is(err, kuma.ErrInvalidTOTPSecret) {
		t.Fatalf("expected kuma.ErrInvalidTOTPSecret, got %v", err)
	}

	if strings.Contains(err.Error(), "pooled") {
		t.Errorf("expected the secret to be rejected before the pool, got %v", err)
	}

	pool := GetGlobalPool()

	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.client != nil || pool.config != nil || pool.refs != 0 {
		t.Errorf("expected an untouched pool, got client %v, config %v, refs %d", pool.client, pool.config, pool.refs)
	}
}

// TestPool_ConfigMatches_Credentials pins the TOTP secret, the session token
// and the auth scheme as part of the pool's connection identity. Two provider
// blocks authenticating as different accounts must not share one connection
// just because neither of them names a username.
func TestPool_ConfigMatches_Credentials(t *testing.T) {
	t.Parallel()

	base := Config{
		Endpoint:     "http://localhost:3001",
		Username:     "admin",
		Password:     "secret",
		TOTPSecret:   "JBSWY3DPEHPK3PXP",
		SessionToken: "token",
		AuthScheme:   kuma.AuthSchemeHTTP,
	}

	pool := &Pool{config: &base}

	same := base
	same.OnSessionTokenRejected = func(error) {}

	if !pool.configMatches(&same) {
		t.Error("expected configMatches to return true for identical credentials and another callback")
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "other session token", mutate: func(c *Config) { c.SessionToken = "other-token" }},
		{name: "missing session token", mutate: func(c *Config) { c.SessionToken = "" }},
		{name: "other totp secret", mutate: func(c *Config) { c.TOTPSecret = "MFRGGZDFMZTWQ2LK" }},
		{name: "missing totp secret", mutate: func(c *Config) { c.TOTPSecret = "" }},
		{name: "other auth scheme", mutate: func(c *Config) { c.AuthScheme = kuma.AuthSchemeSocket }},
		{name: "auto auth scheme", mutate: func(c *Config) { c.AuthScheme = kuma.AuthSchemeAuto }},
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

// TestPool_GetOrCreate_MismatchRedactsSecrets pins that the message explaining
// a rejected config carries no credential. It reaches the user as a Terraform
// diagnostic, in CLI output and in CI logs, and a session token is a bearer
// credential.
func TestPool_GetOrCreate_MismatchRedactsSecrets(t *testing.T) {
	t.Parallel()

	const (
		existingToken  = "existing-session-token"
		requestedToken = "requested-session-token"
		existingSecret = "JBSWY3DPEHPK3PXP"
	)

	pool := &Pool{
		// A non-nil client is what makes GetOrCreate compare configs at all;
		// the mismatch is reported before it is ever used.
		client: &kuma.Client{},
		config: &Config{
			Endpoint:     "http://localhost:3001",
			TOTPSecret:   existingSecret,
			SessionToken: existingToken,
			AuthScheme:   kuma.AuthSchemeHTTP,
		},
	}

	_, err := pool.GetOrCreate(t.Context(), &Config{
		Endpoint:     "http://localhost:3001",
		SessionToken: requestedToken,
	})
	if err == nil {
		t.Fatal("expected a config mismatch error, got nil")
	}

	message := err.Error()

	for _, secret := range []string{existingToken, requestedToken, existingSecret} {
		if strings.Contains(message, secret) {
			t.Errorf("expected the mismatch message not to carry %q, got %q", secret, message)
		}
	}

	for _, want := range []string{
		`existing endpoint="http://localhost:3001" username="" totp_secret=set session_token=set auth_scheme=http`,
		`requested endpoint="http://localhost:3001" username="" totp_secret=unset session_token=set auth_scheme=auto`,
		// Both sides print session_token=set, so only this says it differs.
		`differing secrets: totp_secret, session_token`,
	} {
		if !strings.Contains(message, want) {
			t.Errorf("expected the mismatch message to contain %q, got %q", want, message)
		}
	}
}
