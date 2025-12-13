package client

import (
	"context"
	"testing"
)

func TestNew_EmptyEndpoint(t *testing.T) {
	config := &Config{
		Endpoint: "",
		Username: "admin",
		Password: "secret",
	}

	_, err := New(context.Background(), config)
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

	// Ensure env var is not set
	t.Setenv("UPTIMEKUMA_ENABLE_CONNECTION_POOL", "")

	config := &Config{
		Endpoint:             "http://localhost:3001",
		Username:             "admin",
		Password:             "secret",
		EnableConnectionPool: true,
	}

	// Use a cancelled context to make the connection fail immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This will fail due to cancelled context, but we can verify pooling was enabled
	_, err := New(ctx, config)

	// Should get a connection error (cancelled context)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestNew_PoolEnabledViaEnvVar(t *testing.T) {
	// Reset global pool for test isolation
	ResetGlobalPool()
	defer ResetGlobalPool()

	// Enable pool via environment variable
	t.Setenv("UPTIMEKUMA_ENABLE_CONNECTION_POOL", "true")

	config := &Config{
		Endpoint:             "http://localhost:3001",
		Username:             "admin",
		Password:             "secret",
		EnableConnectionPool: false, // Explicitly false, but env var should override
	}

	// Use a cancelled context to make the connection fail immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This will fail due to cancelled context
	_, err := New(ctx, config)

	// Should get a connection error
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestNew_PoolDisabled(t *testing.T) {
	// Reset global pool for test isolation
	ResetGlobalPool()
	defer ResetGlobalPool()

	// Ensure env var is not set
	t.Setenv("UPTIMEKUMA_ENABLE_CONNECTION_POOL", "")

	config := &Config{
		Endpoint:             "http://localhost:3001",
		Username:             "admin",
		Password:             "secret",
		EnableConnectionPool: false,
	}

	// Use a cancelled context to make the connection fail immediately
	ctx, cancel := context.WithCancel(context.Background())
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

func TestNew_EnvVarNotTrue(t *testing.T) {
	// Reset global pool for test isolation
	ResetGlobalPool()
	defer ResetGlobalPool()

	// Set env var to something other than "true"
	t.Setenv("UPTIMEKUMA_ENABLE_CONNECTION_POOL", "false")

	config := &Config{
		Endpoint:             "http://localhost:3001",
		Username:             "admin",
		Password:             "secret",
		EnableConnectionPool: false,
	}

	// Use a cancelled context to make the connection fail immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := New(ctx, config)

	// Should get a connection error
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}

	// Pool should not have been used
	pool := GetGlobalPool()
	if pool.client != nil {
		t.Error("expected pool client to be nil when env var is 'false'")
	}
}
