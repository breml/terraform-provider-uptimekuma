package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// defaultConnectTimeout is applied when no explicit ConnectTimeout is configured.
// This prevents the provider from hanging indefinitely when Uptime Kuma is unreachable.
const defaultConnectTimeout = 30 * time.Second

// defaultMaxRetries is applied when no explicit MaxRetries is configured.
const defaultMaxRetries = 5

// effectiveTimeout returns the configured timeout, or defaultConnectTimeout if
// the configured value is zero or negative.
func effectiveTimeout(configured time.Duration) time.Duration {
	if configured > 0 {
		return configured
	}

	return defaultConnectTimeout
}

// effectiveMaxRetries returns the configured max retries, or defaultMaxRetries if
// the configured value is zero or negative.
func effectiveMaxRetries(configured int) int {
	if configured > 0 {
		return configured
	}

	return defaultMaxRetries
}

// Config holds the configuration for the Uptime Kuma client.
type Config struct {
	Endpoint             string
	Username             string
	Password             string
	LogLevel             int
	EnableConnectionPool bool
	ConnectTimeout       time.Duration
	MaxRetries           int
}

// New creates a new Uptime Kuma client with optional connection pooling.
// If connection pooling is enabled, it returns a shared connection from the pool.
// Otherwise, it creates a new direct connection with retry logic.
func New(ctx context.Context, config *Config) (*kuma.Client, error) {
	if config.Endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	if config.EnableConnectionPool {
		return GetGlobalPool().GetOrCreate(ctx, config)
	}

	return newClientDirect(ctx, config)
}

// newClientDirect creates a new direct connection with retry logic.
// It resolves the effective timeout (using defaultConnectTimeout when
// none is configured) and bounds each individual connection attempt
// via kuma.WithConnectTimeout. The overall retry process is bounded
// by a separate timer set to ConnectTimeout * (MaxRetries + 1).
// The timer is kept separate from the context passed to kuma.New,
// because the socket.io client stores that context for the lifetime
// of the connection.
func newClientDirect(ctx context.Context, config *Config) (*kuma.Client, error) {
	timeout := effectiveTimeout(config.ConnectTimeout)

	// Shallow copy so we don't mutate the caller's config (important for pool config matching).
	resolved := *config
	resolved.ConnectTimeout = timeout

	opts := []kuma.Option{
		kuma.WithLogLevel(config.LogLevel),
		kuma.WithConnectTimeout(timeout),
	}

	return newClientDirectWithRetry(ctx, &resolved, opts)
}

// newClientDirectWithRetry attempts to connect to Uptime Kuma with
// exponential backoff retry logic. The retry loop is bounded by a
// separate timer set to ConnectTimeout * (MaxRetries + 1), giving
// each attempt a full ConnectTimeout window before the overall
// deadline fires. The timer is intentionally not derived from ctx,
// because ctx is passed into the socket.io client and controls the
// connection lifetime — adding a deadline to it would kill the
// connection after the timeout expires.
func newClientDirectWithRetry(
	ctx context.Context,
	config *Config,
	opts []kuma.Option,
) (*kuma.Client, error) {
	maxRetries := effectiveMaxRetries(config.MaxRetries)

	// Overall deadline = ConnectTimeout * (MaxRetries + 1), so each
	// attempt gets a full ConnectTimeout window before the deadline fires.
	overallDeadline := config.ConnectTimeout * time.Duration(maxRetries+1)
	timer := time.NewTimer(overallDeadline)
	defer timer.Stop()

	deadline := timer.C

	baseDelay := 500 * time.Millisecond

	var kumaClient *kuma.Client
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check overall deadline before each attempt.
		select {
		case <-deadline:
			return nil, newTimeoutError(attempt, err)

		default:
		}

		kumaClient, err = kuma.New(
			ctx,
			config.Endpoint,
			config.Username,
			config.Password,
			opts...,
		)
		if err == nil {
			return kumaClient, nil
		}

		if attempt == maxRetries {
			break
		}

		// Exponential backoff with jitter.
		backoff := float64(baseDelay) * math.Pow(2, float64(attempt))
		//nolint:gosec // Not for cryptographic use, only for jitter in backoff.
		jitter := rand.Float64()*0.4 + 0.8 // 0.8 to 1.2 (±20%)
		sleepDuration := min(time.Duration(backoff*jitter), 30*time.Second)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connection cancelled: %w", ctx.Err())

		case <-deadline:
			return nil, newTimeoutError(attempt+1, err)

		case <-time.After(sleepDuration):
			// Continue retry.
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, err)
}

// newTimeoutError creates a timeout error message. If lastErr is not nil,
// the error includes the number of attempts made and the last error encountered.
func newTimeoutError(attempts int, lastErr error) error {
	if lastErr != nil {
		return fmt.Errorf("connection timed out after %d attempt(s): %w", attempts, lastErr)
	}

	return errors.New("connection timed out")
}
