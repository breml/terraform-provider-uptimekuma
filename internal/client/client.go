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

// Config holds the configuration for the Uptime Kuma client.
type Config struct {
	Endpoint             string
	Username             string
	Password             string
	LogLevel             int
	EnableConnectionPool bool
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
func newClientDirect(ctx context.Context, config *Config) (*kuma.Client, error) {
	maxRetries := 5
	baseDelay := 5 * time.Second

	var kumaClient *kuma.Client
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		kumaClient, err = kuma.New(
			ctx,
			config.Endpoint,
			config.Username,
			config.Password,
			kuma.WithLogLevel(config.LogLevel),
		)
		if err == nil {
			return kumaClient, nil
		}

		if attempt == maxRetries {
			break
		}

		// Exponential backoff with jitter
		backoff := float64(baseDelay) * math.Pow(2, float64(attempt))
		//nolint:gosec // Not for cryptographic use, only for jitter in backoff
		jitter := rand.Float64()*0.4 + 0.8 // 0.8 to 1.2 (±20%)
		sleepDuration := min(time.Duration(backoff*jitter), 30*time.Second)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connection cancelled: %w", ctx.Err())

		case <-time.After(sleepDuration):
			// Continue retry
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, err)
}
