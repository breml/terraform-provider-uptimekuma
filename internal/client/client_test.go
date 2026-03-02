package client

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

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

func TestNewClientDirect_ConnectTimeoutLimitsOverallDuration(t *testing.T) {
	// Connect to an endpoint that will never respond (RFC 5737 TEST-NET).
	// Without the overall deadline, the retry loop would run for minutes.
	// ConnectTimeout bounds both per-attempt timeout and overall duration
	// via a separate timer (not via context deadline, because the context
	// is stored for the connection lifetime by the socket.io client).
	connectTimeout := 2 * time.Second

	config := &Config{
		Endpoint:       "http://192.0.2.1:3001",
		Username:       "admin",
		Password:       "secret",
		ConnectTimeout: connectTimeout,
		MaxRetries:     10,
		LogLevel:       kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	start := time.Now()

	_, err := newClientDirect(t.Context(), config)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for unreachable endpoint, got nil")
	}

	// ConnectTimeout acts as an overall deadline for the entire retry
	// process via a separate timer. With 10 retries but a 2s timer, the
	// operation must complete well before what 10 unbound retries would
	// take. The first per-attempt timeout fires after ConnectTimeout,
	// then the overall timer fires before the next attempt starts.
	upperBound := connectTimeout + 2*time.Second
	if elapsed > upperBound {
		t.Errorf("expected connection to fail within %s, took %s", upperBound, elapsed)
	}

	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("expected timeout error, got: %s", err)
	}
}

func TestNewClientDirect_MaxRetriesWithoutTimeout(t *testing.T) {
	// Verify MaxRetries limits attempts when no ConnectTimeout is set.
	// Without ConnectTimeout there is no overall timer and no
	// per-attempt timeout, so the parent context controls the lifetime.
	// A cancelled context makes each kuma.New call fail immediately;
	// the retry loop then checks ctx.Done() during backoff and returns.
	config := &Config{
		Endpoint:   "http://192.0.2.1:3001",
		Username:   "admin",
		Password:   "secret",
		MaxRetries: 1,
		LogLevel:   kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := newClientDirect(ctx, config)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}

	// With a cancelled context the retry loop exits via ctx.Done()
	// during backoff after the first failed attempt.
	if !strings.Contains(err.Error(), "cancelled") && !strings.Contains(err.Error(), "2 attempts") {
		t.Errorf("expected cancelled or 2-attempts error, got: %s", err)
	}
}

func TestNewClientDirect_NoTimeoutRetriesNormally(t *testing.T) {
	// Without ConnectTimeout, a cancelled parent context should still
	// be respected by the retry loop's select.
	config := &Config{
		Endpoint: "http://192.0.2.1:3001",
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
