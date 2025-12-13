package client

import (
	"os"
	"sync"
	"testing"

	kuma "github.com/breml/go-uptime-kuma-client"
)

func TestPool_RefCount(t *testing.T) {
	pool := &Pool{}

	if pool.RefCount() != 0 {
		t.Errorf("expected ref count 0, got %d", pool.RefCount())
	}

	pool.refs = 5
	if pool.RefCount() != 5 {
		t.Errorf("expected ref count 5, got %d", pool.RefCount())
	}
}

func TestPool_Release(t *testing.T) {
	pool := &Pool{refs: 3}

	pool.Release()
	if pool.RefCount() != 2 {
		t.Errorf("expected ref count 2 after release, got %d", pool.RefCount())
	}

	pool.Release()
	pool.Release()
	if pool.RefCount() != 0 {
		t.Errorf("expected ref count 0, got %d", pool.RefCount())
	}

	// Release should not go negative
	pool.Release()
	if pool.RefCount() != 0 {
		t.Errorf("expected ref count to stay at 0, got %d", pool.RefCount())
	}
}

func TestPool_ConfigMatches(t *testing.T) {
	pool := &Pool{
		config: &Config{
			Endpoint: "http://localhost:3001",
			Username: "admin",
			Password: "secret",
		},
	}

	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name: "matching config",
			config: &Config{
				Endpoint: "http://localhost:3001",
				Username: "admin",
				Password: "secret",
			},
			expected: true,
		},
		{
			name: "different endpoint",
			config: &Config{
				Endpoint: "http://localhost:3002",
				Username: "admin",
				Password: "secret",
			},
			expected: false,
		},
		{
			name: "different username",
			config: &Config{
				Endpoint: "http://localhost:3001",
				Username: "user",
				Password: "secret",
			},
			expected: false,
		},
		{
			name: "different password",
			config: &Config{
				Endpoint: "http://localhost:3001",
				Username: "admin",
				Password: "different",
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.LogLevel = kuma.LogLevel(os.Getenv("SOCKETIO_LOG_LEVEL"))

			result := pool.configMatches(tc.config)
			if result != tc.expected {
				t.Errorf("expected configMatches to return %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestPool_ConfigMatches_NilConfig(t *testing.T) {
	pool := &Pool{config: nil}

	result := pool.configMatches(&Config{Endpoint: "http://localhost:3001"})
	if result {
		t.Error("expected configMatches to return false for nil pool config")
	}
}

func TestPool_Close_NoClient(t *testing.T) {
	pool := &Pool{}

	err := pool.Close()
	if err != nil {
		t.Errorf("expected no error closing empty pool, got %v", err)
	}
}

func TestPool_ConcurrentRelease(t *testing.T) {
	pool := &Pool{refs: 100}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.Release()
		}()
	}
	wg.Wait()

	if pool.RefCount() != 0 {
		t.Errorf("expected ref count 0 after concurrent releases, got %d", pool.RefCount())
	}
}

func TestGetGlobalPool(t *testing.T) {
	// Reset global pool for test isolation
	ResetGlobalPool()

	pool1 := GetGlobalPool()
	pool2 := GetGlobalPool()

	if pool1 != pool2 {
		t.Error("expected GetGlobalPool to return the same instance")
	}

	// Clean up
	ResetGlobalPool()
}

func TestResetGlobalPool(t *testing.T) {
	// Get initial pool
	pool1 := GetGlobalPool()

	// Reset and get new pool
	ResetGlobalPool()
	pool2 := GetGlobalPool()

	if pool1 == pool2 {
		t.Error("expected ResetGlobalPool to create a new pool instance")
	}

	// Clean up
	ResetGlobalPool()
}

func TestCloseGlobalPool_NilPool(t *testing.T) {
	// Reset to ensure globalPool is nil
	ResetGlobalPool()

	err := CloseGlobalPool()
	if err != nil {
		t.Errorf("expected no error closing nil global pool, got %v", err)
	}
}
