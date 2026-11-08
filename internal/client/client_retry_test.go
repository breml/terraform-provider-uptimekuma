package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// errTransport stands in for a transport failure, which is worth retrying.
var errTransport = errors.New("dial tcp: connection refused")

// failingConnect returns a connectFunc that fails every attempt with err and
// counts the attempts.
func failingConnect(err error, attempts *atomic.Int32) connectFunc {
	return func(context.Context, string, string, string, ...kuma.Option) (*kuma.Client, error) {
		attempts.Add(1)
		return nil, err
	}
}

func TestNewClientDirectWithRetry_Attempts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		wantAttempts int32
	}{
		{name: "invalid credentials", err: kuma.ErrInvalidCredentials, wantAttempts: 1},
		{name: "two factor required", err: kuma.ErrTwoFactorRequired, wantAttempts: 1},
		{name: "invalid totp code", err: kuma.ErrInvalidTOTPCode, wantAttempts: 1},
		{name: "two factor locked", err: kuma.ErrTwoFactorLocked, wantAttempts: 1},
		{name: "invalid session token", err: kuma.ErrInvalidSessionToken, wantAttempts: 1},
		{name: "user inactive", err: kuma.ErrUserInactive, wantAttempts: 1},
		{name: "auth required", err: kuma.ErrAuthRequired, wantAttempts: 1},
		{name: "wrapped invalid credentials", err: fmt.Errorf("login: %w", kuma.ErrInvalidCredentials), wantAttempts: 1},
		// Without a RetryAfter the wait is rateLimitWindow, beyond the 30s budget.
		{name: "rate limited without wait", err: kuma.ErrRateLimited, wantAttempts: 1},
		{name: "rate limited with zero wait", err: &kuma.RateLimitedError{}, wantAttempts: 1},
		{
			name:         "rate limited with short wait",
			err:          &kuma.RateLimitedError{RetryAfter: 10 * time.Millisecond},
			wantAttempts: 3,
		},
		{name: "transport error", err: errTransport, wantAttempts: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var attempts atomic.Int32

			config := &Config{
				Endpoint:       "http://uptime-kuma.invalid",
				ConnectTimeout: 30 * time.Second,
				MaxRetries:     2,
			}

			_, err := newClientDirectWithRetry(t.Context(), config, failingConnect(tc.err, &attempts))
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected error wrapping %v, got: %v", tc.err, err)
			}

			if got := attempts.Load(); got != tc.wantAttempts {
				t.Errorf("expected %d attempt(s), got %d", tc.wantAttempts, got)
			}
		})
	}
}

func TestNewClientDirectWithRetry_RateLimitedWaitsRetryAfter(t *testing.T) {
	t.Parallel()

	retryAfter := 100 * time.Millisecond

	var attempts atomic.Int32

	config := &Config{
		Endpoint:       "http://uptime-kuma.invalid",
		ConnectTimeout: 30 * time.Second,
		MaxRetries:     1,
	}

	start := time.Now()

	_, err := newClientDirectWithRetry(
		t.Context(),
		config,
		failingConnect(&kuma.RateLimitedError{RetryAfter: retryAfter}, &attempts),
	)

	elapsed := time.Since(start)

	if !errors.Is(err, kuma.ErrRateLimited) {
		t.Fatalf("expected rate limited error, got: %v", err)
	}

	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}

	// The regular backoff before the second attempt is 400ms to 600ms, so a
	// wait below that shows that RetryAfter replaced it.
	if elapsed < retryAfter || elapsed >= 400*time.Millisecond {
		t.Errorf("expected a wait of about %s, took %s", retryAfter, elapsed)
	}
}

func TestNewClientDirectWithRetry_RateLimitedBeyondBudgetStops(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	config := &Config{
		Endpoint:       "http://uptime-kuma.invalid",
		ConnectTimeout: 5 * time.Second,
		MaxRetries:     3,
	}

	start := time.Now()

	_, err := newClientDirectWithRetry(
		t.Context(),
		config,
		failingConnect(&kuma.RateLimitedError{RetryAfter: time.Minute}, &attempts),
	)

	var rateLimited *kuma.RateLimitedError
	if !errors.As(err, &rateLimited) || rateLimited.RetryAfter != time.Minute {
		t.Fatalf("expected rate limited error with RetryAfter of 1m, got: %v", err)
	}

	if got := attempts.Load(); got != 1 {
		t.Errorf("expected 1 attempt, got %d", got)
	}

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("expected an immediate return, took %s", elapsed)
	}

	if strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected a rate limit error rather than a timeout, got: %v", err)
	}
}

func TestIsPermanentLoginError(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{err: kuma.ErrInvalidCredentials, want: true},
		{err: kuma.ErrTwoFactorRequired, want: true},
		{err: kuma.ErrInvalidTOTPCode, want: true},
		{err: kuma.ErrTwoFactorLocked, want: true},
		{err: kuma.ErrInvalidSessionToken, want: true},
		{err: kuma.ErrUserInactive, want: true},
		{err: kuma.ErrAuthRequired, want: true},
		{err: kuma.ErrRateLimited, want: false},
		{err: &kuma.RateLimitedError{RetryAfter: time.Second}, want: false},
		{err: errTransport, want: false},
		{err: fallbackError(kuma.ErrInvalidCredentials), want: true},
		{err: fallbackError(kuma.ErrTwoFactorRequired), want: true},
		{err: fallbackError(errTransport), want: false},
		{err: fallbackError(context.DeadlineExceeded), want: false},
		{err: fallbackError(&kuma.RateLimitedError{RetryAfter: time.Second}), want: false},
	}

	for _, tc := range tests {
		if got := isPermanentLoginError(tc.err); got != tc.want {
			t.Errorf("isPermanentLoginError(%v) = %t, want %t", tc.err, got, tc.want)
		}
	}
}

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		attempt   int
		remaining time.Duration
		wantMin   time.Duration
		wantMax   time.Duration
		wantRetry bool
	}{
		{
			name:      "backoff first attempt",
			err:       errTransport,
			remaining: time.Minute,
			wantMin:   400 * time.Millisecond,
			wantMax:   600 * time.Millisecond,
			wantRetry: true,
		},
		{
			name:      "backoff third attempt",
			err:       errTransport,
			attempt:   2,
			remaining: time.Minute,
			wantMin:   1600 * time.Millisecond,
			wantMax:   2400 * time.Millisecond,
			wantRetry: true,
		},
		{
			name:      "backoff capped to remaining",
			err:       errTransport,
			attempt:   5,
			remaining: 100 * time.Millisecond,
			wantMin:   100 * time.Millisecond,
			wantMax:   100 * time.Millisecond,
			wantRetry: true,
		},
		{
			name:      "backoff capped to max delay",
			err:       errTransport,
			attempt:   10,
			remaining: time.Hour,
			wantMin:   maxRetryDelay,
			wantMax:   maxRetryDelay,
			wantRetry: true,
		},
		{
			name:      "retry after",
			err:       fmt.Errorf("login: %w", &kuma.RateLimitedError{RetryAfter: 7 * time.Second}),
			remaining: time.Minute,
			wantMin:   7 * time.Second,
			wantMax:   7 * time.Second,
			wantRetry: true,
		},
		{
			name:      "retry after beyond remaining",
			err:       &kuma.RateLimitedError{RetryAfter: 7 * time.Second},
			remaining: 7 * time.Second,
			wantRetry: false,
		},
		{
			name:      "rate limited without retry after waits the window",
			err:       fmt.Errorf("login: %w", kuma.ErrRateLimited),
			remaining: time.Hour,
			wantMin:   rateLimitWindow,
			wantMax:   rateLimitWindow,
			wantRetry: true,
		},
		{
			name:      "rate limited with zero retry after waits the window",
			err:       &kuma.RateLimitedError{},
			remaining: time.Hour,
			wantMin:   rateLimitWindow,
			wantMax:   rateLimitWindow,
			wantRetry: true,
		},
		{
			name:      "rate limited without retry after beyond remaining",
			err:       kuma.ErrRateLimited,
			remaining: defaultConnectTimeout,
			wantRetry: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, retry := retryDelay(tc.err, tc.attempt, tc.remaining)
			if retry != tc.wantRetry {
				t.Fatalf("expected retry %t, got %t", tc.wantRetry, retry)
			}

			if got < tc.wantMin || got > tc.wantMax {
				t.Errorf("expected delay in [%s, %s], got %s", tc.wantMin, tc.wantMax, got)
			}
		})
	}
}

// fallbackError returns the error kuma.New reports for a refused session
// token followed by a password login that failed with passwordErr.
func fallbackError(passwordErr error) error {
	return &kuma.SessionTokenFallbackError{TokenErr: kuma.ErrInvalidSessionToken, PasswordErr: passwordErr}
}

func TestPasswordFallbackCause(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "fallback", err: fallbackError(errTransport), want: errTransport},
		{name: "wrapped fallback", err: fmt.Errorf("login failed: %w", fallbackError(errTransport)), want: errTransport},
		{
			name: "joined with something else",
			err:  errors.Join(errors.New("unrelated"), fallbackError(kuma.ErrInvalidCredentials)),
			want: kuma.ErrInvalidCredentials,
		},
		{name: "plain token rejection", err: kuma.ErrInvalidSessionToken},
		{name: "user inactive", err: kuma.ErrUserInactive},
		{
			name: "two errors, first not a token rejection",
			err:  fmt.Errorf("%w: %w", errTransport, kuma.ErrInvalidCredentials),
		},
		{
			// The shape v0.7.2 reported the fallback in, which is no longer
			// recognised: only the type is.
			name: "untyped fallback shape",
			err:  fmt.Errorf("%w, and the password login that followed: %w", kuma.ErrInvalidSessionToken, errTransport),
		},
		{name: "nil", err: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := PasswordFallbackCause(tc.err)
			if !errors.Is(got, tc.want) || (tc.want == nil) != (got == nil) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// TestNewClientDirectWithRetry_FallbackAttempts pins that a token fallback
// error is retried or not by its password half: it always wraps
// kuma.ErrInvalidSessionToken, which on its own is permanent.
func TestNewClientDirectWithRetry_FallbackAttempts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		wantAttempts int32
	}{
		{name: "password rejected", err: fallbackError(kuma.ErrInvalidCredentials), wantAttempts: 1},
		{name: "two factor required", err: fallbackError(kuma.ErrTwoFactorRequired), wantAttempts: 1},
		{name: "totp code refused", err: fallbackError(kuma.ErrInvalidTOTPCode), wantAttempts: 1},
		{name: "two factor locked", err: fallbackError(kuma.ErrTwoFactorLocked), wantAttempts: 1},
		{name: "transport failure", err: fallbackError(errTransport), wantAttempts: 3},
		{name: "timeout", err: fallbackError(context.DeadlineExceeded), wantAttempts: 3},
		{
			name:         "rate limited with short wait",
			err:          fallbackError(&kuma.RateLimitedError{RetryAfter: 10 * time.Millisecond}),
			wantAttempts: 3,
		},
		// Without a RetryAfter the wait is rateLimitWindow, beyond the budget.
		{name: "rate limited without wait", err: fallbackError(kuma.ErrRateLimited), wantAttempts: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var attempts atomic.Int32

			config := &Config{
				Endpoint:       "http://uptime-kuma.invalid",
				ConnectTimeout: 30 * time.Second,
				MaxRetries:     2,
			}

			_, err := newClientDirectWithRetry(t.Context(), config, failingConnect(tc.err, &attempts))
			if !errors.Is(err, kuma.ErrInvalidSessionToken) {
				t.Fatalf("expected the fallback error, got: %v", err)
			}

			if got := attempts.Load(); got != tc.wantAttempts {
				t.Errorf("expected %d attempt(s), got %d", tc.wantAttempts, got)
			}
		})
	}
}

// TestNewClientDirectWithRetry_FallbackRateLimitedWaitsRetryAfter pins that
// the wait of a rate limit is read out of the password half of a fallback
// error, rather than the regular backoff applied.
func TestNewClientDirectWithRetry_FallbackRateLimitedWaitsRetryAfter(t *testing.T) {
	t.Parallel()

	retryAfter := 100 * time.Millisecond

	var attempts atomic.Int32

	config := &Config{
		Endpoint:       "http://uptime-kuma.invalid",
		ConnectTimeout: 30 * time.Second,
		MaxRetries:     1,
	}

	start := time.Now()

	_, err := newClientDirectWithRetry(
		t.Context(),
		config,
		failingConnect(fallbackError(&kuma.RateLimitedError{RetryAfter: retryAfter}), &attempts),
	)

	elapsed := time.Since(start)

	if !errors.Is(err, kuma.ErrRateLimited) {
		t.Fatalf("expected rate limited error, got: %v", err)
	}

	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}

	// The regular backoff before the second attempt is 400ms to 600ms.
	if elapsed < retryAfter || elapsed >= 400*time.Millisecond {
		t.Errorf("expected a wait of about %s, took %s", retryAfter, elapsed)
	}
}
