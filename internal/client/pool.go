package client

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
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

//nolint:gochecknoglobals // Global pool is a singleton
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
				"pool config mismatch: existing %s, requested %s; differing secrets: %s",
				describeConfig(p.config),
				describeConfig(config),
				differingSecrets(p.config, config),
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

// Release decrements the reference counter for the pooled connection.
// This should be called when a client is no longer needed, but it does not
// actually close the connection (connection remains pooled for reuse).
//
// The provider releases from a goroutine watching the Configure RPC context,
// which is cancelled as soon as Configure returns — long before the client
// stops being used. The counter therefore tracks Configure calls, not live
// users, and the refs check in CloseIfUnused is a leak assertion for TestMain,
// not a use-after-close interlock. Because the release is asynchronous, a
// caller that closes the pool right after its last Configure may still observe
// a non-zero count.
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

	return p.closeLocked()
}

// CloseIfUnused closes the pooled connection, but only when no reference is
// outstanding. The reference check and the close happen in the same critical
// section, so a concurrent Release cannot slip between them.
func (p *Pool) CloseIfUnused() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.refs != 0 {
		return fmt.Errorf("failed to close global pool, expected 0 refs, got: %d", p.refs)
	}

	return p.closeLocked()
}

// closeLocked disconnects and resets the pool. The caller must hold p.mu.
func (p *Pool) closeLocked() error {
	if p.client == nil {
		return nil
	}

	err := p.client.Disconnect()
	p.client = nil
	p.config = nil
	p.refs = 0

	if err != nil {
		return fmt.Errorf("disconnect pooled client: %w", err)
	}

	return nil
}

// describeConfig renders the identity of a configuration for the config
// mismatch message, with every secret reduced to whether it is set, and the
// headers to their names.
func describeConfig(config *Config) string {
	return fmt.Sprintf(
		"endpoint=%q username=%q totp_secret=%s session_token=%s auth_scheme=%s timeout=%s"+
			" per_attempt_timeout=%s operation_timeout=%s max_retries=%d api_key=%s headers=%s user_agent=%q",
		config.Endpoint,
		config.Username,
		secretState(config.TOTPSecret),
		secretState(config.SessionToken),
		config.AuthScheme,
		effectiveTimeout(config.ConnectTimeout),
		config.PerAttemptTimeout,
		effectiveOperationTimeout(config.OperationTimeout),
		effectiveMaxRetries(config.MaxRetries),
		secretState(config.APIKey),
		headerNames(config.Headers),
		config.UserAgent,
	)
}

// headerNames renders the names of the configured headers, sorted, for the
// config mismatch message. The values stay out of it, because a header such as
// CF-Access-Client-Secret is a credential.
func headerNames(headers map[string]string) string {
	if len(headers) == 0 {
		return "[]"
	}

	return "[" + strings.Join(slices.Sorted(maps.Keys(headers)), " ") + "]"
}

// secretState renders whether a credential is configured, for the config
// mismatch message. The values themselves stay out of it, the way the password
// does: the message reaches the user through a Terraform diagnostic.
func secretState(secret string) string {
	if secret == "" {
		return "unset"
	}

	return "set"
}

// differingSecrets names the secrets that differ between two configurations,
// without their values. The mismatch message prints a secret only as set or
// unset, so without this a mismatch in a secret that is set on both sides
// would print two identical halves.
func differingSecrets(existing *Config, requested *Config) string {
	var names []string

	for _, secret := range []struct {
		name                string
		existing, requested string
	}{
		{name: "password", existing: existing.Password, requested: requested.Password},
		{name: "totp_secret", existing: existing.TOTPSecret, requested: requested.TOTPSecret},
		{name: "session_token", existing: existing.SessionToken, requested: requested.SessionToken},
		{name: "api_key", existing: existing.APIKey, requested: requested.APIKey},
	} {
		if secret.existing != secret.requested {
			names = append(names, secret.name)
		}
	}

	// The message names the headers but not their values, so a value that
	// differs under the same name is only told here.
	if !maps.Equal(existing.Headers, requested.Headers) {
		names = append(names, "headers")
	}

	if len(names) == 0 {
		return "none"
	}

	return strings.Join(names, ", ")
}

// configMatches checks if the provided config matches the pool's config.
// Only connection-critical fields (endpoint, credentials including the API
// key, auth scheme, headers, User-Agent, timeout, per_attempt_timeout,
// operation_timeout, max_retries) are compared.
// LogLevel, EnableConnectionPool and OnSessionTokenRejected are intentionally
// excluded as they don't affect the connection identity - the first connection's
// LogLevel is used, and its callback is the only one ever called.
//
// The TOTP secret and the session token are part of that identity as much as
// the password is: two provider blocks authenticating as different accounts
// must not share one connection just because neither names a username. So is
// the API key, which authenticates a connection without any username, and so
// are the headers, which may carry the credentials of an access proxy. The
// User-Agent is compared because a connection is opened with it: a provider
// block naming another one would otherwise be served one it did not ask for.
func (p *Pool) configMatches(config *Config) bool {
	if p.config == nil {
		return false
	}

	return p.config.Endpoint == config.Endpoint &&
		p.config.Username == config.Username &&
		p.config.Password == config.Password &&
		p.config.TOTPSecret == config.TOTPSecret &&
		p.config.SessionToken == config.SessionToken &&
		p.config.AuthScheme == config.AuthScheme &&
		p.config.APIKey == config.APIKey &&
		maps.Equal(p.config.Headers, config.Headers) &&
		p.config.UserAgent == config.UserAgent &&
		effectiveTimeout(p.config.ConnectTimeout) == effectiveTimeout(config.ConnectTimeout) &&
		p.config.PerAttemptTimeout == config.PerAttemptTimeout &&
		effectiveOperationTimeout(p.config.OperationTimeout) ==
			effectiveOperationTimeout(config.OperationTimeout) &&
		effectiveMaxRetries(p.config.MaxRetries) == effectiveMaxRetries(config.MaxRetries)
}

// CloseGlobalPool closes the global connection pool.
// This is a convenience function for test cleanup.
func CloseGlobalPool() error {
	globalPoolMu.Lock()
	defer globalPoolMu.Unlock()

	if globalPool == nil {
		return nil
	}

	return globalPool.CloseIfUnused()
}

// ResetGlobalPool resets the global pool singleton.
// This is primarily for testing purposes to ensure test isolation.
func ResetGlobalPool() {
	globalPoolMu.Lock()
	defer globalPoolMu.Unlock()

	globalPoolOnce = sync.Once{}
	globalPool = nil
}
