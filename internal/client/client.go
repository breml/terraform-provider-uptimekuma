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
// It bounds the overall connection process (across all retry attempts) to
// prevent the provider from hanging indefinitely when Uptime Kuma is unreachable.
const defaultConnectTimeout = 30 * time.Second

// DefaultOperationTimeout is applied when no explicit OperationTimeout is
// configured. It bounds the round trip of a single Uptime Kuma command, so that
// an ack or update event the server never sends fails the operation instead of
// blocking it for as long as the caller's context lives - and Terraform gives a
// provider no deadline of its own.
//
// The upstream client already defaults to the same minute; this constant exists
// so that the provider's own default is explicit, is rendered in the
// `operation_timeout` schema description, and is part of the pool's config
// identity. It is exported for that schema description, so the documented
// default cannot drift from the runtime one.
const DefaultOperationTimeout = 60 * time.Second

// defaultMaxRetries is applied when no explicit MaxRetries is configured.
// It is intentionally small so that the default overall ConnectTimeout budget
// can be split across a few quick attempts without requiring a large total wait.
const defaultMaxRetries = 3

// baseRetryDelay is the backoff before the second attempt; it doubles with
// every further attempt.
const baseRetryDelay = 500 * time.Millisecond

// maxRetryDelay caps the exponential backoff between two attempts.
const maxRetryDelay = 30 * time.Second

// rateLimitWindow is the wait assumed for a rate limited login when the server
// names none. Uptime Kuma 2.x does not, and allows 20 logins per minute, so an
// attempt within the minute is refused again and counts against the limit too.
const rateLimitWindow = time.Minute

// effectiveTimeout returns the configured timeout, or defaultConnectTimeout if
// the configured value is zero or negative.
func effectiveTimeout(configured time.Duration) time.Duration {
	if configured > 0 {
		return configured
	}

	return defaultConnectTimeout
}

// effectiveOperationTimeout returns the configured operation timeout, or
// DefaultOperationTimeout if the configured value is zero or negative. Zero
// means "not configured" here, as it does for ConnectTimeout, and never "no
// bound": kuma.WithOperationTimeout treats a value of zero or less as an
// opt-out that leaves every command bounded only by the caller's context, which
// for a provider is not bounded at all. A caller that wants a very long bound
// configures it explicitly.
func effectiveOperationTimeout(configured time.Duration) time.Duration {
	if configured > 0 {
		return configured
	}

	return DefaultOperationTimeout
}

// effectiveMaxRetries returns the configured max retries, or defaultMaxRetries if
// the configured value is negative. Zero is treated as "no retries" (one attempt total).
func effectiveMaxRetries(configured int) int {
	if configured >= 0 {
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
	// ConnectTimeout is the overall connection budget across all retry
	// attempts. It defaults to defaultConnectTimeout when zero or negative.
	ConnectTimeout time.Duration
	// PerAttemptTimeout, when greater than zero, caps the timeout used for
	// each individual connection attempt. The effective per-attempt timeout
	// is min(PerAttemptTimeout, remainingBudget). When zero, each attempt
	// is allowed to use the full remaining ConnectTimeout budget.
	PerAttemptTimeout time.Duration
	// OperationTimeout bounds each individual command: the wait for the
	// server's ack and for the update event confirming it. It defaults to
	// DefaultOperationTimeout when zero or negative.
	//
	// It is not limited to the operations that follow a connection. The
	// login and setup that ConnectTimeout covers are commands too, so this
	// budget bounds them as well and the effective bound while connecting
	// is whichever of the two expires first. The bound is also on the round
	// trip rather than the whole call: a write that has to queue behind
	// other writes to the list it broadcasts spends no budget waiting its
	// turn.
	OperationTimeout time.Duration
	MaxRetries       int
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
// It resolves the effective overall timeout (using defaultConnectTimeout
// when none is configured) and bounds the entire connection process by
// that budget. Each individual attempt is bounded by
// min(PerAttemptTimeout, remainingBudget) via kuma.WithConnectTimeout.
// The overall budget is enforced by a separate timer, kept independent
// of the context passed to kuma.New, because the socket.io client stores
// that context for the lifetime of the connection.
func newClientDirect(ctx context.Context, config *Config) (*kuma.Client, error) {
	timeout := effectiveTimeout(config.ConnectTimeout)

	// Shallow copy so we don't mutate the caller's config (important for pool config matching).
	resolved := *config
	resolved.ConnectTimeout = timeout

	return newClientDirectWithRetry(ctx, &resolved, kuma.New)
}

// connectFunc opens a single connection to Uptime Kuma. It has the signature of
// kuma.New, which is what the provider uses; tests pass a function that fails
// the way a server would, to exercise the retry decisions without one.
type connectFunc func(
	ctx context.Context,
	baseURL string,
	username string,
	password string,
	opts ...kuma.Option,
) (*kuma.Client, error)

// isPermanentLoginError reports whether err is a login failure that another
// attempt with the same configuration cannot overcome. Retrying those only
// costs time and, worse, login attempts: every one counts against the server's
// login rate limit, so retrying a wrong password can lock the address out for
// a while. kuma.ErrUserInactive wraps kuma.ErrInvalidSessionToken and is
// covered by it.
func isPermanentLoginError(err error) bool {
	for _, target := range []error{
		kuma.ErrInvalidCredentials,
		kuma.ErrTwoFactorRequired,
		kuma.ErrInvalidTOTPCode,
		kuma.ErrTwoFactorLocked,
		kuma.ErrInvalidSessionToken,
		kuma.ErrAuthRequired,
	} {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

// retryDelay returns how long to wait after the failed attempt with the given
// zero-based index before the next one, and false if no further attempt should
// be made. remaining is the rest of the overall budget: it caps the backoff,
// and a rate limit wait that reaches it ends the retries.
//
// A login the server refused for its rate limit is retried only after the wait
// the server named in kuma.RateLimitedError.RetryAfter, or after
// rateLimitWindow where it named none, as Uptime Kuma 2.x does not. An attempt
// before that is refused again and counts against the limit too. A wait at or
// beyond the remaining budget ends the retries at once, so the caller learns
// that it is rate limited rather than that it timed out. With the default
// ConnectTimeout that is always the case for 2.x. A wait just short of the
// budget leaves the next attempt little time, which then ends in a timeout.
// Every other error gets the regular exponential backoff, capped to
// maxRetryDelay.
func retryDelay(err error, attempt int, remaining time.Duration) (time.Duration, bool) {
	if errors.Is(err, kuma.ErrRateLimited) {
		wait := rateLimitWindow

		var rateLimited *kuma.RateLimitedError
		if errors.As(err, &rateLimited) && rateLimited.RetryAfter > 0 {
			wait = rateLimited.RetryAfter
		}

		if wait >= remaining {
			return 0, false
		}

		return wait, true
	}

	// Exponential backoff with jitter, capped to remaining budget.
	backoff := float64(baseRetryDelay) * math.Pow(2, float64(attempt))
	//nolint:gosec // Not for cryptographic use, only for jitter in backoff.
	jitter := rand.Float64()*0.4 + 0.8 // 0.8 to 1.2 (±20%)

	return min(time.Duration(backoff*jitter), maxRetryDelay, remaining), true
}

// newClientDirectWithRetry attempts to connect to Uptime Kuma with
// exponential backoff retry logic. The retry loop is bounded by a
// timer set to the overall ConnectTimeout budget. Each attempt's
// per-attempt timeout is capped to the remaining budget so the
// overall connection process never exceeds ConnectTimeout. The
// timer is intentionally not derived from ctx, because ctx is passed
// into the socket.io client and controls the connection lifetime —
// adding a deadline to it would kill the connection after the
// timeout expires.
//
// A permanent login failure, see isPermanentLoginError, is returned after the
// first attempt, and a rate limited login waits out the rate limit or ends the
// retries, see retryDelay.
func newClientDirectWithRetry(
	ctx context.Context,
	config *Config,
	connect connectFunc,
) (*kuma.Client, error) {
	maxRetries := effectiveMaxRetries(config.MaxRetries)

	overallDeadline := time.Now().Add(config.ConnectTimeout)
	timer := time.NewTimer(time.Until(overallDeadline))
	defer timer.Stop()

	deadline := timer.C

	var kumaClient *kuma.Client
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check overall deadline before each attempt.
		select {
		case <-deadline:
			return nil, newTimeoutError(attempt, err)

		default:
		}

		attemptTimeout := remainingAttemptTimeout(overallDeadline, config.PerAttemptTimeout)
		if attemptTimeout <= 0 {
			return nil, newTimeoutError(attempt, err)
		}

		opts := []kuma.Option{
			kuma.WithLogLevel(config.LogLevel),
			kuma.WithConnectTimeout(attemptTimeout),
			kuma.WithOperationTimeout(effectiveOperationTimeout(config.OperationTimeout)),
		}

		kumaClient, err = connect(
			ctx,
			config.Endpoint,
			config.Username,
			config.Password,
			opts...,
		)
		if err == nil {
			return kumaClient, nil
		}

		if isPermanentLoginError(err) {
			return nil, fmt.Errorf("login failed: %w", err)
		}

		if attempt == maxRetries {
			break
		}

		remaining := time.Until(overallDeadline)
		if remaining <= 0 {
			return nil, newTimeoutError(attempt+1, err)
		}

		sleepDuration, retry := retryDelay(err, attempt, remaining)
		if !retry {
			return nil, fmt.Errorf("failed after %d attempt(s): %w", attempt+1, err)
		}

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

// remainingAttemptTimeout returns the timeout to use for the next attempt.
// It is bounded by the remaining overall budget, and additionally capped to
// perAttempt when perAttempt is greater than zero.
func remainingAttemptTimeout(overallDeadline time.Time, perAttempt time.Duration) time.Duration {
	remaining := time.Until(overallDeadline)
	if remaining <= 0 {
		return 0
	}

	if perAttempt > 0 && perAttempt < remaining {
		return perAttempt
	}

	return remaining
}

// newTimeoutError creates a timeout error message. If lastErr is not nil,
// the error includes the number of attempts made and the last error encountered.
func newTimeoutError(attempts int, lastErr error) error {
	if lastErr != nil {
		return fmt.Errorf("connection timed out after %d attempt(s): %w", attempts, lastErr)
	}

	return errors.New("connection timed out")
}
