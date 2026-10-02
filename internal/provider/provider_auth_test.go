package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

func TestApplyEnvironmentDefaults_Credentials(t *testing.T) {
	t.Setenv("UPTIMEKUMA_TOTP_SECRET", "JBSWY3DPEHPK3PXP")
	t.Setenv("UPTIMEKUMA_SESSION_TOKEN", "token-from-env")

	model := UptimeKumaProviderModel{
		Endpoint:     types.StringNull(),
		Username:     types.StringNull(),
		Password:     types.StringNull(),
		TOTPSecret:   types.StringNull(),
		SessionToken: types.StringNull(),
	}

	applyEnvironmentDefaults(&model, &provider.ConfigureResponse{})

	if model.TOTPSecret.ValueString() != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected totp_secret %q from env, got %q", "JBSWY3DPEHPK3PXP", model.TOTPSecret.ValueString())
	}

	if model.SessionToken.ValueString() != "token-from-env" {
		t.Errorf("expected session_token %q from env, got %q", "token-from-env", model.SessionToken.ValueString())
	}
}

func TestApplyEnvironmentDefaults_CredentialsConfigOverridesEnv(t *testing.T) {
	t.Setenv("UPTIMEKUMA_TOTP_SECRET", "JBSWY3DPEHPK3PXP")
	t.Setenv("UPTIMEKUMA_SESSION_TOKEN", "token-from-env")

	model := UptimeKumaProviderModel{
		Endpoint:     types.StringNull(),
		Username:     types.StringNull(),
		Password:     types.StringNull(),
		TOTPSecret:   types.StringValue("MFRGGZDFMZTWQ2LK"),
		SessionToken: types.StringValue("token-from-config"),
	}

	applyEnvironmentDefaults(&model, &provider.ConfigureResponse{})

	if model.TOTPSecret.ValueString() != "MFRGGZDFMZTWQ2LK" {
		t.Errorf("expected config totp_secret %q to take precedence, got %q",
			"MFRGGZDFMZTWQ2LK", model.TOTPSecret.ValueString())
	}

	if model.SessionToken.ValueString() != "token-from-config" {
		t.Errorf("expected config session_token %q to take precedence, got %q",
			"token-from-config", model.SessionToken.ValueString())
	}
}

// TestValidateCredentials pins the shapes Uptime Kuma can be authenticated
// with. A session token next to a username and password is deliberately
// allowed: the client tries the token first and falls back to the password
// login for a token the server refuses, which is what recovers a configuration
// whose stored token went stale.
func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name         string
		username     types.String
		password     types.String
		totpSecret   types.String
		sessionToken types.String
		wantError    bool
	}{
		{
			name:         "no credentials",
			username:     types.StringNull(),
			password:     types.StringNull(),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringNull(),
		},
		{
			name:         "username and password",
			username:     types.StringValue("admin"),
			password:     types.StringValue("secret"),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringNull(),
		},
		{
			name:         "username, password and totp secret",
			username:     types.StringValue("admin"),
			password:     types.StringValue("secret"),
			totpSecret:   types.StringValue("JBSWY3DPEHPK3PXP"),
			sessionToken: types.StringNull(),
		},
		{
			name:         "session token alone",
			username:     types.StringNull(),
			password:     types.StringNull(),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringValue("token"),
		},
		{
			name:         "session token with username and password",
			username:     types.StringValue("admin"),
			password:     types.StringValue("secret"),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringValue("token"),
		},
		{
			// An attribute fed from an unset variable is not null, but the
			// client is handed nothing, so it has to count as unset.
			name:         "empty totp secret with a session token",
			username:     types.StringNull(),
			password:     types.StringNull(),
			totpSecret:   types.StringValue(""),
			sessionToken: types.StringValue("token"),
		},
		{
			name:         "unknown totp secret with a session token",
			username:     types.StringNull(),
			password:     types.StringNull(),
			totpSecret:   types.StringUnknown(),
			sessionToken: types.StringValue("token"),
		},
		{
			name:         "empty session token is no credential at all",
			username:     types.StringNull(),
			password:     types.StringNull(),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringValue(""),
		},
		{
			// The client rejects this pairing itself, with an error that
			// carries no sentinel and arrives looking like a connection
			// failure, so it is worth catching here.
			name:         "empty username with a password",
			username:     types.StringValue(""),
			password:     types.StringValue("secret"),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringNull(),
			wantError:    true,
		},
		{
			name:         "username without password",
			username:     types.StringValue("admin"),
			password:     types.StringNull(),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringNull(),
			wantError:    true,
		},
		{
			name:         "password without username",
			username:     types.StringNull(),
			password:     types.StringValue("secret"),
			totpSecret:   types.StringNull(),
			sessionToken: types.StringNull(),
			wantError:    true,
		},
		{
			name:         "totp secret alone",
			username:     types.StringNull(),
			password:     types.StringNull(),
			totpSecret:   types.StringValue("JBSWY3DPEHPK3PXP"),
			sessionToken: types.StringNull(),
			wantError:    true,
		},
		{
			name:         "totp secret with session token only",
			username:     types.StringNull(),
			password:     types.StringNull(),
			totpSecret:   types.StringValue("JBSWY3DPEHPK3PXP"),
			sessionToken: types.StringValue("token"),
			wantError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := UptimeKumaProviderModel{
				Endpoint:     types.StringValue("http://localhost:3001"),
				Username:     tc.username,
				Password:     tc.password,
				TOTPSecret:   tc.totpSecret,
				SessionToken: tc.sessionToken,
			}

			resp := &provider.ConfigureResponse{}

			validateCredentials(&model, resp)

			if got := resp.Diagnostics.HasError(); got != tc.wantError {
				t.Errorf("expected error %t, got %t (%v)", tc.wantError, got, resp.Diagnostics.Errors())
			}
		})
	}
}

// TestConnectionErrorDiagnostic pins that a login the server refused is named
// as such instead of arriving as the generic connection failure every cause
// used to share. The sentinels are wrapped the way the pool wraps them on the
// way out, because a classification that only works on a bare error would
// never fire in practice.
func TestConnectionErrorDiagnostic(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantSummary string
	}{
		{name: "auth required", err: kuma.ErrAuthRequired, wantSummary: "credentials required"},
		{name: "invalid credentials", err: kuma.ErrInvalidCredentials, wantSummary: "invalid credentials"},
		{
			name:        "two-factor required",
			err:         kuma.ErrTwoFactorRequired,
			wantSummary: "two-factor authentication required",
		},
		{
			name:        "invalid one-time code",
			err:         kuma.ErrInvalidTOTPCode,
			wantSummary: "invalid two-factor authentication code",
		},
		{
			name:        "user inactive wins over the token rejection it wraps",
			err:         kuma.ErrUserInactive,
			wantSummary: "user inactive or deleted",
		},
		{
			name:        "session token rejected",
			err:         kuma.ErrInvalidSessionToken,
			wantSummary: "session token rejected",
		},
		{name: "rate limited", err: kuma.ErrRateLimited, wantSummary: "too many login attempts"},
		{
			name:        "malformed totp secret",
			err:         client.ErrInvalidTOTPSecret,
			wantSummary: "invalid totp_secret",
		},
		{
			name:        "anything else stays the connection failure",
			err:         errors.New("connect to server: context deadline exceeded"),
			wantSummary: "failed to connect to Uptime Kuma",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf("create pooled connection: %w", tc.err)

			summary, detail := connectionErrorDiagnostic("http://localhost:3001", wrapped)

			if summary != tc.wantSummary {
				t.Errorf("expected summary %q, got %q", tc.wantSummary, summary)
			}

			if !strings.Contains(detail, tc.err.Error()) {
				t.Errorf("expected detail to carry the underlying error %q, got %q", tc.err, detail)
			}
		})
	}
}

// TestAccProviderTOTPSecretWithoutCredentials asserts the rejection of a
// configuration the provider could not honour: the one-time code a
// `totp_secret` produces is only ever asked for by a password login.
func TestAccProviderTOTPSecretWithoutCredentials(t *testing.T) {
	// applyEnvironmentDefaults runs before validateCredentials, so a username
	// and password left in the environment would turn this into the valid
	// shape and the configuration would be rejected by the server it then
	// dials, not by the validation under test.
	t.Setenv("UPTIMEKUMA_USERNAME", "")
	t.Setenv("UPTIMEKUMA_PASSWORD", "")

	// Not resource.ParallelTest: t.Setenv above marks this test as one that
	// cannot run in parallel, and t.Parallel would panic and take the whole
	// test binary down with it.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "uptimekuma" {
  endpoint      = "http://localhost:3001"
  session_token = "token"
  totp_secret   = "JBSWY3DPEHPK3PXP"
}

data "uptimekuma_tag" "test" {}
`,
				ExpectError: regexp.MustCompile(`totp_secret`),
			},
		},
	})
}

// TestAccProviderMalformedTOTPSecret asserts that a secret no code can be
// derived from is named at plan time, by the attribute, rather than reaching
// the server and being reported as a connection failure.
func TestAccProviderMalformedTOTPSecret(t *testing.T) {
	t.Setenv("UPTIMEKUMA_USERNAME", "")
	t.Setenv("UPTIMEKUMA_PASSWORD", "")

	// Not resource.ParallelTest, see TestAccProviderTOTPSecretWithoutCredentials.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "uptimekuma" {
  endpoint    = "http://localhost:3001"
  username    = "admin"
  password    = "password"
  totp_secret = "JBSWY3DPEHPK3PX0"
}

data "uptimekuma_tag" "test" {}
`,
				ExpectError: regexp.MustCompile(`base32 shared secret`),
			},
		},
	})
}

// TestAccSessionTokenLogin connects with the token of a login that already
// happened, which is the only test that exercises SessionToken end to end -
// from the provider's attribute through client.Config to kuma.WithSessionToken.
//
// It goes through client.New rather than a Terraform configuration because the
// pool's config identity now includes the credentials: a provider block with a
// different credential shape than providerConfig() is refused by the pool
// before it ever reaches the server.
func TestAccSessionTokenLogin(t *testing.T) {
	if endpoint == "" {
		t.Skip("acceptance tests not enabled - skipping session token login test")
	}

	token := outOfBandClient.SessionToken()
	if token == "" {
		t.Fatal("expected the out-of-band client to hold a session token")
	}

	tokenClient, err := client.New(t.Context(), &client.Config{
		Endpoint:     endpoint,
		SessionToken: token,
	})
	if err != nil {
		t.Fatalf("expected the session token to authenticate, got %v", err)
	}

	defer func() {
		disconnectErr := tokenClient.Disconnect()
		if disconnectErr != nil {
			t.Errorf("failed to disconnect the token client: %v", disconnectErr)
		}
	}()

	// A read the server only answers for a logged-in client, so the assertion
	// is that the token authenticated and not merely that a socket opened.
	_, err = tokenClient.GetMonitors(t.Context())
	if err != nil {
		t.Errorf("expected a token authenticated client to read monitors, got %v", err)
	}
}

// TestWarnRejectedSessionToken pins the one report that a stored token has
// died. The connection succeeds without it, so a dropped call or an inverted
// condition would leave the configuration applying until the day the password
// beside it is removed.
func TestWarnRejectedSessionToken(t *testing.T) {
	accepted := &provider.ConfigureResponse{}

	warnRejectedSessionToken(nil, accepted)

	if count := accepted.Diagnostics.WarningsCount(); count != 0 {
		t.Errorf("expected no warning for an accepted token, got %d", count)
	}

	rejection := errors.New("login by token: user inactive or deleted")
	rejected := &provider.ConfigureResponse{}

	warnRejectedSessionToken(rejection, rejected)

	if rejected.Diagnostics.HasError() {
		t.Errorf("expected a rejected token to warn rather than fail, got %v", rejected.Diagnostics.Errors())
	}

	warnings := rejected.Diagnostics.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected exactly one warning, got %d", len(warnings))
	}

	if summary := warnings[0].Summary(); summary != "session_token rejected" {
		t.Errorf("expected summary %q, got %q", "session_token rejected", summary)
	}

	if detail := warnings[0].Detail(); !strings.Contains(detail, rejection.Error()) {
		t.Errorf("expected the warning to carry the reason %q, got %q", rejection, detail)
	}
}

// TestPasswordFallbackDiagnostic pins the classification of the error the
// client returns when a refused session token was followed by a password login
// that failed too. It wraps both sentinels, so reporting it by the token
// rejection alone would bury the half the user has to act on - and, for the
// rate limit, the only half that resolves on its own.
func TestPasswordFallbackDiagnostic(t *testing.T) {
	tests := []struct {
		name        string
		followUp    error
		wantSummary string
	}{
		{
			name:        "rate limited",
			followUp:    kuma.ErrRateLimited,
			wantSummary: "session_token rejected, password login rate limited",
		},
		{
			name:        "two-factor required",
			followUp:    kuma.ErrTwoFactorRequired,
			wantSummary: "session_token rejected, password login needs a one-time code",
		},
		{
			name:        "one-time code refused",
			followUp:    kuma.ErrInvalidTOTPCode,
			wantSummary: "session_token rejected, one-time code refused",
		},
		{
			name:        "password refused",
			followUp:    kuma.ErrInvalidCredentials,
			wantSummary: "session_token and password both rejected",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The shape the client returns, see authenticate in
			// github.com/breml/go-uptime-kuma-client.
			compound := fmt.Errorf(
				"create pooled connection: %w, and the password login that followed: %w",
				kuma.ErrInvalidSessionToken,
				tc.followUp,
			)

			summary, detail := connectionErrorDiagnostic("http://localhost:3001", compound)

			if summary != tc.wantSummary {
				t.Errorf("expected summary %q, got %q", tc.wantSummary, summary)
			}

			if !strings.Contains(detail, tc.followUp.Error()) {
				t.Errorf("expected detail to carry the follow-up error %q, got %q", tc.followUp, detail)
			}
		})
	}
}

// TestPasswordFallbackDiagnostic_NotACompoundError pins that a token rejection
// on its own keeps its own diagnostic. kuma.ErrUserInactive is the case the
// client never follows with a password login, and it wraps
// kuma.ErrInvalidSessionToken, so it would otherwise look compound.
func TestPasswordFallbackDiagnostic_NotACompoundError(t *testing.T) {
	for _, err := range []error{kuma.ErrInvalidSessionToken, kuma.ErrUserInactive} {
		if summary, _ := passwordFallbackDiagnostic(err); summary != "" {
			t.Errorf("expected no compound diagnostic for %v, got %q", err, summary)
		}
	}
}

// TestAccTerminalAuthErrorIsNotRetried pins the behaviour the named login
// diagnostics rest on: a credential the server refused is reported as such,
// once, instead of being repeated until the attempts run out. Without it the
// user gets "failed after 4 attempts" in place of the reason, and four of the
// server's twenty logins per minute are spent on a password that cannot
// improve.
func TestAccTerminalAuthErrorIsNotRetried(t *testing.T) {
	if endpoint == "" {
		t.Skip("acceptance tests not enabled - skipping terminal auth error test")
	}

	_, err := client.New(t.Context(), &client.Config{
		Endpoint: endpoint,
		Username: username,
		Password: "not-the-password",
	})
	if !errors.Is(err, kuma.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if !strings.HasPrefix(err.Error(), "authenticate:") {
		t.Errorf("expected the rejection to be reported as an authentication failure, got %v", err)
	}

	if strings.Contains(err.Error(), "failed after") {
		t.Errorf("expected the rejected credential not to be retried, got %v", err)
	}
}
