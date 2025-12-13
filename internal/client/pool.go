package client

import (
	"context"
	"fmt"
	"sync"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// Pool manages a shared connection to Uptime Kuma for testing scenarios.
// This prevents "login: Too frequently" errors during acceptance tests by
// reusing a single Socket.IO connection across multiple provider instances.
type Pool struct {
	mu     sync.Mutex
	client *kuma.Client
	config *Config
	refs   int
}

var (
	globalPool     *Pool
	globalPoolOnce sync.Once
	globalPoolMu   sync.Mutex // protects globalPool and globalPoolOnce during reset
)

// GetGlobalPool returns the global connection pool singleton.
func GetGlobalPool() *Pool {
	globalPoolMu.Lock()
	defer globalPoolMu.Unlock()
	globalPoolOnce.Do(func() {
		globalPool = &Pool{}
	})
	return globalPool
}

// GetOrCreate returns an existing client from the pool or creates a new one.
// If a client already exists with different configuration, an error is returned
// to prevent credential confusion.
func (p *Pool) GetOrCreate(ctx context.Context, config *Config) (*kuma.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		if !p.configMatches(config) {
			return nil, fmt.Errorf(
				"pool config mismatch: existing endpoint=%q username=%q, requested endpoint=%q username=%q",
				p.config.Endpoint, p.config.Username, config.Endpoint, config.Username,
			)
		}
		p.refs++
		return p.client, nil
	}

	client, err := newClientDirect(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pooled connection: %w", err)
	}

	p.client = client
	p.config = config
	p.refs = 1

	return client, nil
}

// configMatches checks if the provided config matches the pool's config.
// Only connection-critical fields (endpoint, credentials) are compared.
// LogLevel and EnableConnectionPool are intentionally excluded as they don't
// affect the connection identity - the first connection's LogLevel is used.
func (p *Pool) configMatches(config *Config) bool {
	if p.config == nil {
		return false
	}
	return p.config.Endpoint == config.Endpoint &&
		p.config.Username == config.Username &&
		p.config.Password == config.Password
}

// Release decrements the reference counter for the pooled connection.
// This should be called when a client is no longer needed, but it does not
// actually close the connection (connection remains pooled for reuse).
//
// Note: In the current acceptance test use case, Release is not called by
// consumers because the pool is closed via CloseGlobalPool at the end of all
// tests. The reference count is maintained for debugging purposes and to
// support future use cases where automatic cleanup when refs reach zero
// might be desired.
func (p *Pool) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.refs > 0 {
		p.refs--
	}
}

// RefCount returns the current reference count (for testing/debugging).
func (p *Pool) RefCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.refs
}

// Close forcefully closes the pooled connection and resets the pool.
// This should only be called during test cleanup (e.g., in TestMain).
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		err := p.client.Disconnect()
		p.client = nil
		p.config = nil
		p.refs = 0
		return err
	}
	return nil
}

// CloseGlobalPool closes the global connection pool.
// This is a convenience function for test cleanup.
func CloseGlobalPool() error {
	if globalPool != nil {
		return globalPool.Close()
	}
	return nil
}

// ResetGlobalPool resets the global pool singleton.
// This is primarily for testing purposes to ensure test isolation.
func ResetGlobalPool() {
	globalPoolMu.Lock()
	defer globalPoolMu.Unlock()
	globalPoolOnce = sync.Once{}
	globalPool = nil
}
