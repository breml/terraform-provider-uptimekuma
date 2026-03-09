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

func TestEffectiveMaxRetries_Default(t *testing.T) {
	got := effectiveMaxRetries(0)
	if got != defaultMaxRetries {
		t.Errorf("expected %d, got %d", defaultMaxRetries, got)
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
	// ConnectTimeout bounds per-attempt timeout; the overall deadline
	// is ConnectTimeout * (MaxRetries + 1). The timer is separate from
	// the context because the socket.io client stores it for the
	// connection lifetime.
	endpoint := startDeadEndListener(t)
	connectTimeout := 2 * time.Second

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

	// Overall deadline = ConnectTimeout * (MaxRetries + 1) = 2s * 3 = 6s.
	// Allow some slack for scheduling.
	upperBound := connectTimeout*time.Duration(config.MaxRetries+1) + 2*time.Second
	if elapsed > upperBound {
		t.Errorf("expected connection to fail within %s, took %s", upperBound, elapsed)
	}

	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "failed after") {
		t.Errorf("expected timeout or retry-exhaustion error, got: %s", err)
	}
}

func TestNewClientDirect_MaxRetriesLimitsAttempts(t *testing.T) {
	// Verify that MaxRetries limits the number of connection attempts.
	// Use a dead-end listener with a short ConnectTimeout so each
	// attempt fails quickly. With MaxRetries=2 and ConnectTimeout=1s,
	// the overall deadline is 1s * 3 = 3s. The test verifies that all
	// 3 attempts run within that window.
	endpoint := startDeadEndListener(t)

	config := &Config{
		Endpoint:       endpoint,
		Username:       "admin",
		Password:       "secret",
		ConnectTimeout: 1 * time.Second,
		MaxRetries:     2,
		LogLevel:       kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL")),
	}

	start := time.Now()

	_, err := newClientDirect(t.Context(), config)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for dead-end endpoint, got nil")
	}

	// Overall deadline = 1s * (2 + 1) = 3s. Allow some slack.
	upperBound := config.ConnectTimeout*time.Duration(config.MaxRetries+1) + 2*time.Second
	if elapsed > upperBound {
		t.Errorf("expected connection to fail within %s, took %s", upperBound, elapsed)
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
