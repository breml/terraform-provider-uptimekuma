package client

import (
	"context"
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
