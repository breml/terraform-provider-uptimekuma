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
		secrets     requestSecrets
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
			// The client also reports an expired 3.x two-factor step with
			// totp_secret set, so the advice must cover a configured secret too.
			wantDetail: []string{"TOTP secret", "If `totp_secret`", "is not set", "expired", "`session_token`"},
			notDetail:  []string{"issues/404"},
		},
		{
			name:        "invalid totp code",
			err:         kuma.ErrInvalidTOTPCode,
			wantSummary: "Uptime Kuma rejected the two-factor authentication code",
			wantDetail:  []string{"TOTP secret", "`totp_secret`", "30 second step", "retries once", "`session_token`"},
			notDetail:   []string{"issues/404", "per_attempt_timeout"},
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
			// On 3.x the two-factor step can expire with totp_secret set, so
			// the advice must not claim that no secret is configured.
			wantDetail: []string{"fell back to failed as well", "Set `totp_secret` if it is not set", "expired"},
			notDetail:  []string{"no TOTP secret is configured"},
		},
		{
			name:        "session token rejected, totp code refused",
			err:         passwordFallbackError(kuma.ErrInvalidTOTPCode),
			wantSummary: "Uptime Kuma rejected the session token and the two-factor authentication code",
			wantDetail: []string{
				"fell back to failed as well", "check `totp_secret`", "Uptime Kuma 2.x", "retry in the next step",
			},
		},
		{
			name:        "session token rejected, two factor locked",
			err:         passwordFallbackError(kuma.ErrTwoFactorLocked),
			wantSummary: "Uptime Kuma rejected the session token and locked the two-factor authentication",
			wantDetail:  []string{"fell back to failed as well", "too many wrong ones"},
		},
		{
			// Uptime Kuma 3.x reports a banned account on the password login
			// that follows a lapsed session, so the password half can be
			// inactive even though the token half never is.
			name:        "session token rejected, account inactive",
			err:         passwordFallbackError(kuma.ErrUserInactive),
			wantSummary: "Uptime Kuma rejected the session token and the account is inactive",
			wantDetail:  []string{"fell back to failed as well", "deactivated or deleted"},
			notDetail:   []string{"does not recognize", "Common causes"},
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
		{
			name:        "api key on a 2.x server",
			err:         fmt.Errorf("create pooled connection: login failed: %w", kuma.ErrAPIKeyUnsupported),
			wantSummary: "Uptime Kuma does not accept API keys for this connection",
			wantDetail:  []string{endpoint, "only on Uptime Kuma 3.0", "this server is Uptime Kuma 2.x"},
			notDetail:   []string{"Common causes", "`auth_scheme`"},
		},
		{
			name:        "api key rejected",
			err:         fmt.Errorf("create pooled connection: login failed: %w", kuma.ErrAPIKeyRejected),
			wantSummary: "Uptime Kuma rejected the API key",
			wantDetail: []string{
				"invalid, expired or revoked", `"API and MCP" is disabled in Settings`,
				"`auth_scheme = \"http\"` is pinned against an Uptime Kuma 2.x server",
			},
		},
		{
			name:        "api key against a server that is not set up",
			err:         fmt.Errorf("login failed: %w", kuma.ErrAPIKeySetupRequired),
			wantSummary: "Uptime Kuma is not set up yet",
			wantDetail:  []string{"username and password first", "`api_key`"},
		},
		{
			name: "rejected header",
			err: errors.New("create pooled connection: login failed: " +
				"headers: Host is set by the client or its transport and cannot be configured"),
			wantSummary: "Invalid api_key, headers or user_agent",
			wantDetail:  []string{endpoint, "Host is set by the client", "`Authorization: Bearer`"},
			notDetail:   []string{"refused the login", "Common causes"},
		},
		{
			name:        "rejected api key format",
			err:         errors.New("api key: pass the key without the Bearer prefix"),
			wantSummary: "Invalid api_key, headers or user_agent",
			notDetail:   []string{"trimspace"},
		},
		{
			name:        "api key with whitespace",
			err:         errors.New("api key: the key contains whitespace or control characters"),
			wantSummary: "Invalid api_key, headers or user_agent",
			wantDetail:  []string{"use trimspace() when reading the key from a file"},
		},
		{
			name: "userinfo next to an api key",
			err: errors.New("base url: userinfo (user:password@) sets an Authorization header of its own, " +
				"which cannot be combined with WithAPIKey or an Authorization header from WithHeaders"),
			wantSummary: "endpoint contains a user name and password",
			wantDetail:  []string{"Remove the user:password@ part from `endpoint`"},
		},
		{
			name: "redirect to https",
			err: errors.New(`create pooled connection: login failed: Get "http://kuma.example.com/api/entry-page": ` +
				"redirect from http://kuma.example.com/api/entry-page to https://kuma.example.com/api/entry-page " +
				"upgrades to https; use https://kuma.example.com as the base URL"),
			wantSummary: "Uptime Kuma redirected to https",
			wantDetail:  []string{"use https://kuma.example.com as the base URL", "to the https URL", "into GET"},
			notDetail:   []string{"refused the login", "Common causes", "API key", "cloudflareaccess"},
		},
		{
			name:        "redirect to another host with headers",
			err:         errors.New(crossOriginRedirect),
			secrets:     withRequestSecrets,
			wantSummary: "Uptime Kuma redirected to another origin",
			wantDetail:  []string{"cloudflareaccess.com", "service token in `headers`", "Otherwise set `endpoint`"},
		},
		{
			name: "redirect to another host without headers",
			err: errors.New(`Get "https://kuma.example.com/api/entry-page": redirect from ` +
				"https://kuma.example.com/api/entry-page to https://status.example.com/api/entry-page leaves " +
				"the server's origin; use https://status.example.com as the base URL"),
			wantSummary: "Uptime Kuma redirected to another origin",
			wantDetail:  []string{"into GET", "Set `endpoint`"},
			notDetail:   []string{"cloudflareaccess", "API key", "`headers`"},
		},
		{
			name: "auth scheme probe answered by a proxy",
			err: errors.New(`connect to server: detect auth scheme: GET http://kuma.example.com/api/auth/ok ` +
				`answered status 403 with Content-Type "text/html", which is neither Uptime Kuma 2.x nor 3.x; ` +
				"an access proxy in front of the server may be rejecting the request (check WithHeaders), or " +
				"pin the scheme with WithAuthScheme"),
			wantSummary: "Unexpected answer to the auth scheme detection",
			wantDetail:  []string{"`auth_scheme = \"socket\"`", "up to 0.6.0", "`headers`"},
			notDetail:   []string{"service token"},
		},
		{
			name:        "auth required names api_key",
			err:         kuma.ErrAuthRequired,
			wantSummary: "Uptime Kuma requires authentication",
			wantDetail:  []string{"`api_key`"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary, detail := connectionError(endpoint, tc.err, tc.secrets)
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

// crossOriginRedirect is the error the client refuses a redirect to another
// host with.
const crossOriginRedirect = `create pooled connection: login failed: Get "https://kuma.example.com/api/entry-page": ` +
	"redirect from https://kuma.example.com/api/entry-page to https://team.cloudflareaccess.com/login " +
	"leaves the server's origin; connect to the server on https://team.cloudflareaccess.com"

// TestConnectionError_RedactsEndpoint pins that the password of an endpoint
// with userinfo never reaches a diagnostic.
func TestConnectionError_RedactsEndpoint(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		errors.New("dial tcp: connection refused"),
		kuma.ErrInvalidCredentials,
		errors.New("base url: userinfo (user:password@) sets an Authorization header of its own"),
		errors.New(crossOriginRedirect),
	} {
		summary, detail := connectionError("https://admin:s3cr3t@kuma.example.com", err, withRequestSecrets)
		if strings.Contains(summary+detail, "s3cr3t") {
			t.Errorf("expected the password redacted for %q, got:\n%s", err, detail)
		}

		if !strings.Contains(detail, "admin:xxxxx@kuma.example.com") {
			t.Errorf("expected the redacted endpoint for %q, got:\n%s", err, detail)
		}
	}
}
