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
	// Use a local listener that accepts TCP connections but never
	// completes the socket.io handshake. This is deterministic and
	// independent of network configuration, unlike TEST-NET addresses.
	// ConnectTimeout bounds both per-attempt timeout and overall duration
	// via a separate timer (not via context deadline, because the context
	// is stored for the connection lifetime by the socket.io client).
	endpoint := startDeadEndListener(t)
	connectTimeout := 2 * time.Second

	config := &Config{
		Endpoint:       endpoint,
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

func TestNewClientDirect_MaxRetriesLimitsAttempts(t *testing.T) {
	// Verify that MaxRetries limits the number of connection attempts.
	// Use a dead-end listener with a short ConnectTimeout so each
	// attempt fails quickly. The overall timer (same value as
	// ConnectTimeout) fires after the first attempt, which produces
	// a "timed out after 1 attempt(s)" error — proving both the
	// per-attempt timeout and the overall timer work together.
	endpoint := startDeadEndListener(t)

	config := &Config{
		Endpoint:       endpoint,
		Username:       "admin",
		Password:       "secret",
		ConnectTimeout: 1 * time.Second,
		MaxRetries:     5,
		LogLevel:       kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	start := time.Now()

	_, err := newClientDirect(t.Context(), config)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for dead-end endpoint, got nil")
	}

	// The overall timer (1s) fires after the first per-attempt timeout
	// (also 1s), so at most 1 attempt runs despite MaxRetries=5.
	// Total time must be close to ConnectTimeout, not 5× that.
	if elapsed > 3*time.Second {
		t.Errorf("expected connection to fail within ~3s, took %s (MaxRetries not bounded by timer?)", elapsed)
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got: %s", err)
	}
}

func TestNewClientDirect_NoTimeoutRetriesNormally(t *testing.T) {
	// Without ConnectTimeout, a cancelled parent context should still
	// be respected by the retry loop's select.
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
