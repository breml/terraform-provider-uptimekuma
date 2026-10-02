package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// startDeadEndListener starts a TCP listener that accepts connections but
// never sends any data. This provides a deterministic, fast-failing
// endpoint for tests: the TCP handshake succeeds immediately, but the
// socket.io handshake never completes, so kuma.New blocks until its
// per-attempt ConnectTimeout fires. Returns the listener address.
func startDeadEndListener(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start dead-end listener: %v", err)
	}

	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			// Hold the connection open without sending anything.
			go func() {
				<-time.After(30 * time.Second)
				conn.Close()
			}()
		}
	}()

	return fmt.Sprintf("http://%s", ln.Addr().String())
}

func TestEffectiveTimeout_Default(t *testing.T) {
	got := effectiveTimeout(0)
	if got != defaultConnectTimeout {
		t.Errorf("expected %s, got %s", defaultConnectTimeout, got)
	}
}

func TestEffectiveTimeout_Explicit(t *testing.T) {
	explicit := 10 * time.Second

	got := effectiveTimeout(explicit)
	if got != explicit {
		t.Errorf("expected %s, got %s", explicit, got)
	}
}

func TestEffectiveTimeout_Negative(t *testing.T) {
	got := effectiveTimeout(-5 * time.Second)
	if got != defaultConnectTimeout {
		t.Errorf("expected %s for negative input, got %s", defaultConnectTimeout, got)
	}
}

// TestEffectiveOperationTimeout covers the bound on a single Uptime Kuma
// operation. Unlike max retries, zero here is "not configured" and not "no
// bound": an unset attribute must still leave a command bounded, because an
// unbounded one blocks an apply for as long as Terraform lets it, which is
// forever.
func TestEffectiveOperationTimeout(t *testing.T) {
	tests := []struct {
		name       string
		configured time.Duration
		want       time.Duration
	}{
		{name: "unset falls back to the default", configured: 0, want: DefaultOperationTimeout},
		{name: "negative falls back to the default", configured: -5 * time.Second, want: DefaultOperationTimeout},
		{name: "explicit value is kept", configured: 90 * time.Second, want: 90 * time.Second},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := effectiveOperationTimeout(tc.configured)
			if got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestEffectiveMaxRetries_Default(t *testing.T) {
	got := effectiveMaxRetries(-1)
	if got != defaultMaxRetries {
		t.Errorf("expected %d, got %d", defaultMaxRetries, got)
	}
}

func TestEffectiveMaxRetries_Zero(t *testing.T) {
	got := effectiveMaxRetries(0)
	if got != 0 {
		t.Errorf("expected 0 for explicitly-zero max retries, got %d", got)
	}
}

func TestEffectiveMaxRetries_Explicit(t *testing.T) {
	got := effectiveMaxRetries(10)
	if got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
}

func TestEffectiveMaxRetries_Negative(t *testing.T) {
	got := effectiveMaxRetries(-3)
	if got != defaultMaxRetries {
		t.Errorf("expected %d for negative input, got %d", defaultMaxRetries, got)
	}
}

func TestRemainingAttemptTimeout_NoCap(t *testing.T) {
	deadline := time.Now().Add(5 * time.Second)

	got := remainingAttemptTimeout(deadline, 0)
	if got <= 0 || got > 5*time.Second {
		t.Errorf("expected value within (0,5s], got %s", got)
	}
}

func TestRemainingAttemptTimeout_CapApplied(t *testing.T) {
	deadline := time.Now().Add(10 * time.Second)
	perAttempt := 2 * time.Second

	got := remainingAttemptTimeout(deadline, perAttempt)
	if got != perAttempt {
		t.Errorf("expected %s, got %s", perAttempt, got)
	}
}

func TestRemainingAttemptTimeout_CapLargerThanRemaining(t *testing.T) {
	deadline := time.Now().Add(1 * time.Second)
	perAttempt := 30 * time.Second

	got := remainingAttemptTimeout(deadline, perAttempt)
	if got <= 0 || got > 1*time.Second {
		t.Errorf("expected value within (0,1s], got %s", got)
	}
}

func TestRemainingAttemptTimeout_DeadlinePassed(t *testing.T) {
	deadline := time.Now().Add(-1 * time.Second)

	got := remainingAttemptTimeout(deadline, 5*time.Second)
	if got != 0 {
		t.Errorf("expected 0 when deadline already passed, got %s", got)
	}
}

func TestNew_EmptyEndpoint(t *testing.T) {
	config := &Config{
		Endpoint: "",
		Username: "admin",
		Password: "secret",
		LogLevel: kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	_, err := New(t.Context(), config)
	if err == nil {
		t.Error("expected error for empty endpoint, got nil")
	}

	expectedMsg := "endpoint is required"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestNew_PoolEnabledViaConfig(t *testing.T) {
	// Reset global pool for test isolation
	ResetGlobalPool()
	defer ResetGlobalPool()

	config := &Config{
		Endpoint:             "http://localhost:3001",
		Username:             "admin",
		Password:             "secret",
		EnableConnectionPool: true,
		LogLevel:             kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	// Use a cancelled context to make the connection fail immediately
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	// This will fail due to cancelled context, but we can verify pooling was enabled
	_, err := New(ctx, config)

	// Should get a connection error (cancelled context)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestNew_PoolDisabled(t *testing.T) {
	// Reset global pool for test isolation
	ResetGlobalPool()
	defer ResetGlobalPool()

	config := &Config{
		Endpoint:             "http://localhost:3001",
		Username:             "admin",
		Password:             "secret",
		EnableConnectionPool: false,
		LogLevel:             kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	// Use a cancelled context to make the connection fail immediately
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := New(ctx, config)

	// Should get a connection cancelled error
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}

	// Pool should not have been used (client is nil in pool)
	pool := GetGlobalPool()
	if pool.client != nil {
		t.Error("expected pool client to be nil when pooling disabled")
	}
}

// attemptsFromError reads the attempt count out of the error newTimeoutError
// builds. Elapsed time alone cannot tell one long attempt from several short
// ones, so tests that care about retry behaviour assert on this instead.
func attemptsFromError(t *testing.T, err error) int {
	t.Helper()

	var attempts int

	_, scanErr := fmt.Sscanf(err.Error(), "connection timed out after %d attempt(s)", &attempts)
	if scanErr != nil {
		t.Fatalf("could not read the attempt count from %q: %v", err, scanErr)
	}

	return attempts
}

func TestNewClientDirect_ConnectTimeoutLimitsOverallDuration(t *testing.T) {
	// Use a local listener that accepts TCP connections but never
	// completes the socket.io handshake. This is deterministic and
	// independent of network configuration, unlike TEST-NET addresses.
	// The timer is separate from the context because the socket.io client
	// stores it for the connection lifetime.
	//
	// With no PerAttemptTimeout the single attempt is capped at the whole
	// remaining budget, so this pins the total wall-clock bound and no retry
	// can start. TestNewClientDirect_PerAttemptTimeoutLimitsAttempts is the
	// same budget with a per-attempt cap, and gets a retry out of it.
	endpoint := startDeadEndListener(t)
	connectTimeout := time.Second

	config := &Config{
		Endpoint:       endpoint,
		Username:       "admin",
		Password:       "secret",
		ConnectTimeout: connectTimeout,
		MaxRetries:     2,
		LogLevel:       kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	start := time.Now()

	_, err := newClientDirect(t.Context(), config)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for unreachable endpoint, got nil")
	}

	// Overall deadline equals ConnectTimeout (total budget). Allow some
	// slack for scheduling and for the in-flight kuma.New attempt to
	// observe the deadline.
	upperBound := connectTimeout + time.Second
	if elapsed > upperBound {
		t.Errorf("expected connection to fail within %s, took %s", upperBound, elapsed)
	}

	// The budget has to be spent, not returned early, or the bound above
	// would pass for a connection that never waited at all.
	if elapsed < connectTimeout {
		t.Errorf("expected the full %s budget to be used, took only %s", connectTimeout, elapsed)
	}

	if attempts := attemptsFromError(t, err); attempts != 1 {
		t.Errorf("expected the uncapped attempt to consume the whole budget, got %d attempts", attempts)
	}

	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "failed after") {
		t.Errorf("expected timeout or retry-exhaustion error, got: %s", err)
	}
}

func TestNewClientDirect_PerAttemptTimeoutLimitsAttempts(t *testing.T) {
	// Verify that PerAttemptTimeout caps each individual attempt so that a
	// retry fits inside a budget one uncapped attempt would swallow whole.
	//
	// The budget has to clear the first backoff or the cap buys nothing:
	// attempt 0 spends 100ms, the backoff is baseDelay (500ms) with up to
	// +20% jitter, so a second attempt starts by 700ms at the latest. A 1s
	// budget therefore always gets two attempts, and never a third, which
	// would need another 800-1200ms of backoff.
	endpoint := startDeadEndListener(t)

	config := &Config{
		Endpoint:          endpoint,
		Username:          "admin",
		Password:          "secret",
		ConnectTimeout:    time.Second,
		PerAttemptTimeout: 100 * time.Millisecond,
		MaxRetries:        5,
		LogLevel:          kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	start := time.Now()

	_, err := newClientDirect(t.Context(), config)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for dead-end endpoint, got nil")
	}

	// Overall deadline equals ConnectTimeout. Allow some slack.
	upperBound := config.ConnectTimeout + time.Second
	if elapsed > upperBound {
		t.Errorf("expected connection to fail within %s, took %s", upperBound, elapsed)
	}

	if attempts := attemptsFromError(t, err); attempts < 2 {
		t.Errorf("expected the per-attempt cap to allow a retry, got %d attempt(s)", attempts)
	}

	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "failed after") {
		t.Errorf("expected timeout or retry-exhaustion error, got: %s", err)
	}
}

func TestNewClientDirect_CancelledContextReturnsError(t *testing.T) {
	// A cancelled parent context should be respected by the retry loop's
	// select, even when using the default timeout.
	endpoint := startDeadEndListener(t)

	config := &Config{
		Endpoint: endpoint,
		Username: "admin",
		Password: "secret",
		LogLevel: kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := newClientDirect(ctx, config)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

// TestTerminalAuthError pins which failures the retry loop must not repeat.
// Every attempt costs one of the 20 logins per minute Uptime Kuma allows - two
// for one that answers a one-time code - and a rejected credential is rejected
// just as firmly on the fourth try, with the retry count replacing the reason
// the server gave.
func TestTerminalAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "auth required", err: kuma.ErrAuthRequired, want: true},
		{name: "invalid credentials", err: kuma.ErrInvalidCredentials, want: true},
		{name: "two-factor required", err: kuma.ErrTwoFactorRequired, want: true},
		{name: "invalid one-time code", err: kuma.ErrInvalidTOTPCode, want: true},
		{name: "session token rejected", err: kuma.ErrInvalidSessionToken, want: true},
		{name: "user inactive", err: kuma.ErrUserInactive, want: true},
		{
			// The limiter is a bucket that refills, so the retry loop is the
			// mechanism that recovers from it, not an attempt wasted.
			name: "rate limited is worth a retry",
			err:  kuma.ErrRateLimited,
			want: false,
		},
		{
			name: "wrapped sentinel",
			err:  fmt.Errorf("login: %w: the server asked for a code again", kuma.ErrTwoFactorRequired),
			want: true,
		},
		{name: "transport failure is worth a retry", err: errors.New("connect to server: EOF"), want: false},
		{name: "no error", err: nil, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := terminalAuthError(tc.err); got != tc.want {
				t.Errorf("expected terminalAuthError %t, got %t", tc.want, got)
			}
		})
	}
}

// TestConnectOptions pins that each credential option only appears once its
// own field is configured, so a client without them keeps the plain password
// login - and, for the TOTP secret, so that kuma.New is never handed the empty
// secret it rejects outright.
//
// A kuma.Option can only be applied to a kuma.Client, whose fields this
// package cannot read, so the options are counted rather than identified. The
// count is per field, which is what keeps one guard from covering for the
// other; TestNewClientDirect_SessionTokenIsNotTheTOTPSecret pins that the two
// are not swapped.
func TestConnectOptions(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   int
	}{
		{name: "no credentials", config: &Config{}, want: 3},
		{name: "totp secret only", config: &Config{TOTPSecret: "JBSWY3DPEHPK3PXP"}, want: 4},
		{name: "session token only", config: &Config{SessionToken: "token"}, want: 4},
		{
			name:   "rejection callback only",
			config: &Config{OnSessionTokenRejected: func(error) {}},
			want:   4,
		},
		{
			name:   "all of them",
			config: &Config{TOTPSecret: "JBSWY3DPEHPK3PXP", SessionToken: "token", OnSessionTokenRejected: func(error) {}},
			want:   6,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(connectOptions(tc.config, time.Second)); got != tc.want {
				t.Errorf("expected %d options, got %d", tc.want, got)
			}
		})
	}
}

// TestValidateTOTPSecret pins the tolerances the upstream client documents for
// a secret copied out of Uptime Kuma's two-factor dialog, and the rejection
// that keeps a typo from being retried as an unreachable server.
func TestValidateTOTPSecret(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		wantError bool
	}{
		{name: "base32", secret: "JBSWY3DPEHPK3PXP"},
		{name: "lower case", secret: "jbswy3dpehpk3pxp"},
		{name: "spaces and hyphens", secret: "jbsw y3dp-ehpk 3pxp"},
		{name: "missing padding", secret: "JBSWY3DPEHPK3PX"},
		{name: "empty", secret: "", wantError: true},
		{name: "only separators", secret: " - = ", wantError: true},
		{name: "not base32", secret: "JBSWY3DPEHPK3PX0", wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTOTPSecret(tc.secret)

			if got := err != nil; got != tc.wantError {
				t.Fatalf("expected error %t, got %v", tc.wantError, err)
			}

			if tc.wantError && !errors.Is(err, ErrInvalidTOTPSecret) {
				t.Errorf("expected error to match ErrInvalidTOTPSecret, got %v", err)
			}
		})
	}
}

// TestNew_MalformedTOTPSecretIsNotRetried pins that a secret no code can be
// derived from fails before the first connection attempt. Retrying it would
// spend the server's login budget on a value that cannot improve, and report
// the result as a connection failure.
func TestNew_MalformedTOTPSecretIsNotRetried(t *testing.T) {
	ctx := t.Context()

	start := time.Now()

	_, err := New(ctx, &Config{Endpoint: "http://127.0.0.1:1", TOTPSecret: "JBSWY3DPEHPK3PX0"})
	if !errors.Is(err, ErrInvalidTOTPSecret) {
		t.Fatalf("expected ErrInvalidTOTPSecret, got %v", err)
	}

	// The first backoff alone is 500ms, so anything near it means the secret
	// went down the retry path.
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("expected the secret to be rejected before connecting, took %s", elapsed)
	}
}

// TestNewClientDirect_SessionTokenIsNotTheTOTPSecret pins that the session
// token is not handed to kuma.WithTOTPSecret. A token is not base32, so a
// swapped pair of options fails the connection with a secret decoding error
// instead of the transport error a refused port produces.
func TestNewClientDirect_SessionTokenIsNotTheTOTPSecret(t *testing.T) {
	ctx := t.Context()

	_, err := New(ctx, &Config{
		Endpoint:     "http://127.0.0.1:1",
		SessionToken: "not-a-base32-token!",
		MaxRetries:   0,
	})
	if err == nil {
		t.Fatal("expected an error connecting to a refused port, got nil")
	}

	if strings.Contains(err.Error(), "totp") {
		t.Errorf("expected the session token not to be read as a totp secret, got %v", err)
	}
}
