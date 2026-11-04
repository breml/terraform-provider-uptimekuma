package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	kuma "github.com/breml/go-uptime-kuma-client"
)

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
			wantDetail:  []string{"issues/404"},
		},
		{
			name:        "invalid totp code",
			err:         kuma.ErrInvalidTOTPCode,
			wantSummary: "Uptime Kuma rejected the two-factor authentication code",
			wantDetail:  []string{"issues/404"},
			notDetail:   []string{"TOTP secret"},
		},
		{
			name:        "two factor locked",
			err:         kuma.ErrTwoFactorLocked,
			wantSummary: "Uptime Kuma locked the two-factor authentication",
			wantDetail:  []string{"too many wrong ones", "issues/404"},
			notDetail:   []string{"TOTP secret"},
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
			wantDetail:  []string{"password of the account was changed", "`username` and `password`"},
			notDetail:   []string{"fallback"},
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
			wantDetail:  []string{"no credentials are configured"},
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
