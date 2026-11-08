package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// passwordFallbackError returns the error the client reports when a refused
// session token was followed by a password login that failed with followUp,
// wrapped the way the provider receives it.
func passwordFallbackError(followUp error) error {
	return fmt.Errorf(
		"create pooled connection: login failed: %w",
		&kuma.SessionTokenFallbackError{TokenErr: kuma.ErrInvalidSessionToken, PasswordErr: followUp},
	)
}

func TestConnectionError(t *testing.T) {
	const endpoint = "https://uptime-kuma.example.com"

	tests := []struct {
		name        string
		err         error
		wantSummary string
		wantDetail  []string
		notDetail   []string
	}{
		{
			name:        "invalid credentials",
			err:         fmt.Errorf("create pooled connection: login failed: %w", kuma.ErrInvalidCredentials),
			wantSummary: "Uptime Kuma rejected the username or password",
			wantDetail:  []string{endpoint, kuma.ErrInvalidCredentials.Error(), "`username` and `password`"},
			notDetail:   []string{"Common causes"},
		},
		{
			name:        "two factor required",
			err:         kuma.ErrTwoFactorRequired,
			wantSummary: "Uptime Kuma requires two-factor authentication",
			wantDetail:  []string{"TOTP secret", "`totp_secret`", "`session_token`"},
			notDetail:   []string{"issues/404"},
		},
		{
			name:        "invalid totp code",
			err:         kuma.ErrInvalidTOTPCode,
			wantSummary: "Uptime Kuma rejected the two-factor authentication code",
			wantDetail:  []string{"TOTP secret", "`totp_secret`", "30 second step", "`session_token`"},
			notDetail:   []string{"issues/404"},
		},
		{
			name:        "two factor locked",
			err:         kuma.ErrTwoFactorLocked,
			wantSummary: "Uptime Kuma locked the two-factor authentication",
			wantDetail:  []string{"too many wrong ones", "TOTP secret", "`totp_secret`"},
			notDetail:   []string{"issues/404"},
		},
		{
			name:        "user inactive",
			err:         kuma.ErrUserInactive,
			wantSummary: "Uptime Kuma account is inactive",
			wantDetail:  []string{"deactivated or deleted"},
		},
		{
			name:        "invalid session token",
			err:         kuma.ErrInvalidSessionToken,
			wantSummary: "Uptime Kuma rejected the session token",
			wantDetail: []string{
				"`session_token`", "password of the account was changed", "7 days", "`username` and `password`",
			},
			notDetail: []string{"fell back"},
		},
		{
			name:        "rate limited with wait",
			err:         fmt.Errorf("failed after 1 attempt(s): %w", &kuma.RateLimitedError{RetryAfter: 8 * time.Second}),
			wantSummary: "Uptime Kuma rate limited the login",
			wantDetail:  []string{"Wait 8s"},
		},
		{
			name:        "rate limited without wait",
			err:         kuma.ErrRateLimited,
			wantSummary: "Uptime Kuma rate limited the login",
			wantDetail:  []string{"Wait a minute"},
		},
		{
			name:        "rate limited with zero wait",
			err:         fmt.Errorf("failed after 1 attempt(s): %w", &kuma.RateLimitedError{}),
			wantSummary: "Uptime Kuma rate limited the login",
			wantDetail:  []string{"Wait a minute"},
			notDetail:   []string{"Wait 0s"},
		},
		{
			name:        "auth required",
			err:         kuma.ErrAuthRequired,
			wantSummary: "Uptime Kuma requires authentication",
			wantDetail:  []string{"no credentials are configured", "`session_token`"},
		},
		{
			name: "malformed totp secret",
			// client.New checks the secret before it reaches the pool, so the
			// error carries no pool wrapping.
			err:         fmt.Errorf("%w: empty", kuma.ErrInvalidTOTPSecret),
			wantSummary: "Invalid TOTP secret",
			wantDetail:  []string{"`totp_secret`", "UPTIMEKUMA_TOTP_SECRET", "base32", "empty"},
			notDetail:   []string{"refused the login", "Common causes"},
		},
		{
			name:        "session token rejected, password rejected",
			err:         passwordFallbackError(kuma.ErrInvalidCredentials),
			wantSummary: "Uptime Kuma rejected the session token and the password",
			wantDetail: []string{
				"fell back to failed as well",
				"check `password` first",
				kuma.ErrInvalidCredentials.Error(),
			},
		},
		{
			name:        "session token rejected, two factor required",
			err:         passwordFallbackError(kuma.ErrTwoFactorRequired),
			wantSummary: "Uptime Kuma rejected the session token and requires two-factor authentication",
			wantDetail:  []string{"fell back to failed as well", "Set `totp_secret`"},
		},
		{
			name:        "session token rejected, totp code refused",
			err:         passwordFallbackError(kuma.ErrInvalidTOTPCode),
			wantSummary: "Uptime Kuma rejected the session token and the two-factor authentication code",
			wantDetail:  []string{"fell back to failed as well", "Check `totp_secret`"},
		},
		{
			name:        "session token rejected, two factor locked",
			err:         passwordFallbackError(kuma.ErrTwoFactorLocked),
			wantSummary: "Uptime Kuma rejected the session token and the two-factor authentication code",
			wantDetail:  []string{"fell back to failed as well"},
		},
		{
			name:        "session token rejected, rate limited",
			err:         passwordFallbackError(&kuma.RateLimitedError{RetryAfter: 8 * time.Second}),
			wantSummary: "Uptime Kuma rejected the session token and rate limited the login",
			wantDetail:  []string{"fell back to failed as well", "Wait 8s", "retry after 8s"},
			notDetail:   []string{"Wait a minute"},
		},
		{
			name:        "session token rejected, rate limited without wait",
			err:         passwordFallbackError(kuma.ErrRateLimited),
			wantSummary: "Uptime Kuma rejected the session token and rate limited the login",
			wantDetail:  []string{"fell back to failed as well", "Wait a minute"},
		},
		{
			// The client never falls back for an inactive user, but the
			// sentinel wraps kuma.ErrInvalidSessionToken, so it must not be
			// mistaken for the compound error.
			name:        "user inactive is not a fallback",
			err:         fmt.Errorf("login failed: %w", kuma.ErrUserInactive),
			wantSummary: "Uptime Kuma account is inactive",
		},
		{
			// A password login that failed for a reason without a sentinel
			// still names the fallback, and must not tell the user to set the
			// username and password they already have.
			name:        "session token rejected, untyped follow-up",
			err:         fmt.Errorf("failed after 4 attempts: %w", passwordFallbackError(errors.New("connection reset"))),
			wantSummary: "Uptime Kuma rejected the session token and the password login failed",
			wantDetail: []string{
				"fell back to failed as well",
				"does not recognize",
				"connection reset",
				"SOCKETIO_LOG_LEVEL",
			},
			notDetail: []string{"next to it", "for the provider to fall back on"},
		},
		{
			// A token rejection without a fallback, as when no password is
			// configured, keeps the token diagnostic.
			name:        "session token rejected without fallback",
			err:         fmt.Errorf("create pooled connection: login failed: %w", kuma.ErrInvalidSessionToken),
			wantSummary: "Uptime Kuma rejected the session token",
			wantDetail:  []string{"for the provider to fall back on"},
		},
		{
			name:        "transport error",
			err:         errors.New("dial tcp: connection refused"),
			wantSummary: "failed to connect to Uptime Kuma",
			wantDetail: []string{
				endpoint, "connection refused", "Common causes", "does not recognize", "SOCKETIO_LOG_LEVEL",
			},
			notDetail: []string{"Two-factor", "2FA"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary, detail := connectionError(endpoint, tc.err)
			if summary != tc.wantSummary {
				t.Errorf("expected summary %q, got %q", tc.wantSummary, summary)
			}

			for _, want := range tc.wantDetail {
				if !strings.Contains(detail, want) {
					t.Errorf("expected detail to contain %q, got:\n%s", want, detail)
				}
			}

			for _, notWant := range tc.notDetail {
				if strings.Contains(detail, notWant) {
					t.Errorf("expected detail not to contain %q, got:\n%s", notWant, detail)
				}
			}
		})
	}
}
