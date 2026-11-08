package provider

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// The values of the `auth_scheme` attribute. They are the names
// kuma.AuthScheme.String gives the schemes, so the documentation, the
// validator and the pool mismatch message all spell them the same way.
const (
	authSchemeAuto   = "auto"
	authSchemeSocket = "socket"
	authSchemeHTTP   = "http"
)

// sessionTokenDescription documents the `session_token` attribute. It lives
// here rather than inline so that the schema stays readable.
const sessionTokenDescription = "Session token from an earlier login, used instead of the password. " +
	"It bypasses two-factor authentication, so it needs neither the password nor a one-time code. " +
	"On Uptime Kuma 2.x it is the JWT the server hands out on every login; it does not expire, and " +
	"only a password change or a deactivated account invalidates it. On Uptime Kuma 3.x it is the " +
	"value of the `better-auth.session_token` cookie, which lapses after 7 days without use and on a " +
	"password change, so keep `username` and `password` next to it there. Set it alone, or together " +
	"with `username` and `password`: the token is tried first, and the password login is the fallback " +
	"for a token the server refuses, which the provider reports with a warning. `uptimekuma_settings` " +
	"needs `password` on the provider, because Uptime Kuma asks for the account password again when " +
	"settings are written. Can be set via `UPTIMEKUMA_SESSION_TOKEN` environment variable, which " +
	"applies when the attribute is unset or empty."

// totpSecretDescription documents the `totp_secret` attribute.
const totpSecretDescription = "Shared secret of an account with two-factor authentication enabled, " +
	"for the password login. This is the base32 `secret` from the `otpauth://` URI Uptime Kuma shows " +
	"while two-factor authentication is set up; spaces, hyphens, lower case and missing padding are " +
	"accepted. The provider derives the one-time code the server asks for from it, and only once the " +
	"server asks, so it is inert on an account without two-factor authentication. Requires `username` " +
	"and `password`. Uptime Kuma 2.x refuses a code it has already accepted for the account, and every " +
	"Terraform command logs in on its own: `terraform apply` logs in twice, for the plan and for the " +
	"apply, and the second login often falls into the 30 second step of the first. The provider then " +
	"waits for the next step, so set `timeout` to at least 60s, and leave `per_attempt_timeout` unset " +
	"or above 35s. The same applies to parallel runs, where the single retry settles " +
	"two logins in one step but not more; set `session_token` next to the password to need no code. " +
	"Can be set via `UPTIMEKUMA_TOTP_SECRET` environment variable, which applies when the attribute is " +
	"unset or empty."

// authSchemeDescription documents the `auth_scheme` attribute. The names
// match authSchemeAuto, authSchemeSocket and authSchemeHTTP.
const authSchemeDescription = "How the provider authenticates: `auto` (default) detects it with one " +
	"request before connecting, `socket` is the login command on the Socket.IO connection of Uptime " +
	"Kuma 2.x, and `http` is the HTTP sign-in of Uptime Kuma 3.x. The detection is right for every " +
	"Uptime Kuma release; pin the scheme only for a reverse proxy that answers the detection request " +
	"itself. A pinned scheme the server does not speak fails the login. Can be set via " +
	"`UPTIMEKUMA_AUTH_SCHEME` environment variable, which applies when the attribute is unset or " +
	"empty."

// parseAuthScheme maps an `auth_scheme` value to the client's scheme, and
// reports whether the value is one of the names. An empty value, which is what
// an unset attribute reads as, is the detection. This is the one table of the
// names: the schema validator, the check of the environment value and the
// client configuration all go through it, so an unknown name can never become
// the detection silently.
func parseAuthScheme(value string) (kuma.AuthScheme, bool) {
	switch value {
	case "", authSchemeAuto:
		return kuma.AuthSchemeAuto, true

	case authSchemeSocket:
		return kuma.AuthSchemeSocket, true

	case authSchemeHTTP:
		return kuma.AuthSchemeHTTP, true

	default:
		return kuma.AuthSchemeAuto, false
	}
}

// applyCredentialEnvironmentDefaults applies the environment defaults of the
// credentials and the auth scheme.
//
// Unlike the other attributes in applyEnvironmentDefaults, these consult the
// environment when the value is null or empty, not only when it is null: an
// empty attribute counts as unset,
// so an attribute fed from an unset variable reaches the environment rather
// than shadowing it with an empty string. An unknown value is left alone,
// because it is not unset, merely not known yet; Configure rejects it.
func applyCredentialEnvironmentDefaults(data *UptimeKumaProviderModel) {
	data.Username = stringFromEnvironment(data.Username, "UPTIMEKUMA_USERNAME")
	data.Password = stringFromEnvironment(data.Password, "UPTIMEKUMA_PASSWORD")
	data.SessionToken = stringFromEnvironment(data.SessionToken, "UPTIMEKUMA_SESSION_TOKEN")
	data.TOTPSecret = stringFromEnvironment(data.TOTPSecret, "UPTIMEKUMA_TOTP_SECRET")
	data.AuthScheme = stringFromEnvironment(data.AuthScheme, "UPTIMEKUMA_AUTH_SCHEME")
}

// stringFromEnvironment returns value, or the value of the environment variable
// name where value is null or empty and the variable is set. An unknown value
// is returned unchanged.
func stringFromEnvironment(value types.String, name string) types.String {
	if value.IsUnknown() || value.ValueString() != "" {
		return value
	}

	env := os.Getenv(name)
	if env == "" {
		return value
	}

	return types.StringValue(env)
}

// rejectUnknownCredentials reports every credential attribute, and the auth
// scheme, that is still unknown when the provider is configured. That happens
// when the provider configuration refers to a value Terraform only learns
// during apply, such as an attribute of a resource that does not exist yet.
// The provider has to log in to plan at all, and an unknown value would reach
// the client as an empty string: a login as nobody, failing with an error that
// names the wrong cause.
func rejectUnknownCredentials(data *UptimeKumaProviderModel, diags *diag.Diagnostics) {
	for _, attribute := range []struct {
		name  string
		value types.String
	}{
		{name: "username", value: data.Username},
		{name: "password", value: data.Password},
		{name: "session_token", value: data.SessionToken},
		{name: "totp_secret", value: data.TOTPSecret},
		{name: "auth_scheme", value: data.AuthScheme},
	} {
		if !attribute.value.IsUnknown() {
			continue
		}

		diags.AddAttributeError(
			path.Root(attribute.name),
			fmt.Sprintf("unknown %s", attribute.name),
			fmt.Sprintf(
				"The provider attribute %s refers to a value that is not known until apply, but the "+
					"provider logs in to Uptime Kuma during plan, so its credentials must be known at plan "+
					"time. Set %s from a variable, a data source or an environment variable instead of from "+
					"a resource that is created in the same run.",
				attribute.name, attribute.name,
			),
		)
	}
}

// ValidateConfig checks the values the schema validators cannot see, because
// they come from the environment: an `auth_scheme` from UPTIMEKUMA_AUTH_SCHEME
// and a `totp_secret` from UPTIMEKUMA_TOTP_SECRET. A malformed value is wrong
// whatever else is set, so `terraform validate` reports it.
//
// The credential shape is deliberately not judged here but only in Configure.
// Every shape error is a credential that is missing, and the environment may
// supply it: a configuration with a literal `username` whose password comes
// from UPTIMEKUMA_PASSWORD only in the apply job is valid, yet a `terraform
// validate` run without that secret would reject it. Configure runs before
// the provider connects, so the shape is still reported before any login.
func (*UptimeKumaProvider) ValidateConfig(
	ctx context.Context,
	req provider.ValidateConfigRequest,
	resp *provider.ValidateConfigResponse,
) {
	var data UptimeKumaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// A configured value has been checked by the schema validator, so only one
	// taken from the environment still needs checking.
	authSchemeFromEnvironment := data.AuthScheme.IsNull() || data.AuthScheme.ValueString() == ""
	totpSecretFromEnvironment := data.TOTPSecret.IsNull() || data.TOTPSecret.ValueString() == ""

	applyCredentialEnvironmentDefaults(&data)

	if authSchemeFromEnvironment {
		resolveAuthScheme(&data, &resp.Diagnostics)
	}

	if totpSecretFromEnvironment {
		validateEnvironmentTOTPSecret(&data, &resp.Diagnostics)
	}
}

// validateEnvironmentTOTPSecret rejects a `totp_secret` taken from
// UPTIMEKUMA_TOTP_SECRET that is not a base32 shared secret. The schema
// validator only sees the configuration; without this, the environment value
// would only be caught by client.New during Configure, after `terraform
// validate` had passed. An unset value passes.
func validateEnvironmentTOTPSecret(data *UptimeKumaProviderModel, diags *diag.Diagnostics) {
	secret := data.TOTPSecret.ValueString()
	if data.TOTPSecret.IsUnknown() || secret == "" {
		return
	}

	err := kuma.ValidateTOTPSecret(secret)
	if err == nil {
		return
	}

	diags.AddAttributeError(
		path.Root("totp_secret"),
		"Invalid TOTP secret",
		fmt.Sprintf(
			"The environment variable UPTIMEKUMA_TOTP_SECRET is not a base32 shared secret, so no "+
				"one-time code can be derived from it. Copy the `secret` out of the `otpauth://` URI "+
				"Uptime Kuma showed while two-factor authentication was set up.\n\nUnderlying error: %v",
			err,
		),
	)
}

// resolveAuthScheme returns the client's scheme for `auth_scheme`, and
// rejects a value that is not one of the names. The schema validator only sees
// the configuration, so without this a typo in UPTIMEKUMA_AUTH_SCHEME would
// silently fall back to the detection. A null or unknown value resolves to the
// detection without an error; Configure rejects the unknown one itself.
func resolveAuthScheme(data *UptimeKumaProviderModel, diags *diag.Diagnostics) kuma.AuthScheme {
	if data.AuthScheme.IsUnknown() {
		return kuma.AuthSchemeAuto
	}

	scheme, ok := parseAuthScheme(data.AuthScheme.ValueString())
	if !ok {
		diags.AddAttributeError(
			path.Root("auth_scheme"),
			"invalid auth_scheme",
			fmt.Sprintf(
				"auth_scheme (or the environment variable UPTIMEKUMA_AUTH_SCHEME) must be one of %q, %q or "+
					"%q, got %q",
				authSchemeAuto, authSchemeSocket, authSchemeHTTP, data.AuthScheme.ValueString(),
			),
		)
	}

	return scheme
}

// validateCredentials checks that the configured credentials form one of the
// shapes Uptime Kuma can be authenticated with: nothing at all, for a server
// with authentication disabled; a username and password, optionally with a
// `totp_secret` for an account that has two-factor authentication enabled;
// or a `session_token`, on its own or with a username and password to fall
// back on.
//
// A `totp_secret` without a username and password is rejected rather than
// ignored: only a password login is ever asked for a one-time code, so the
// secret could not take effect and its presence says the configuration means
// something it does not do.
//
// The shapes are judged by the value the client will be handed, not by whether
// the attribute is null: `totp_secret = var.secret` with an empty default is a
// value the client ignores, and rejecting it would fail a configuration that
// works. A credential that is still unknown cannot be judged at all, so the
// check is skipped; Configure, its only caller, rejects an unknown credential
// before it gets here.
func validateCredentials(data *UptimeKumaProviderModel, diags *diag.Diagnostics) {
	for _, value := range []types.String{data.Username, data.Password, data.SessionToken, data.TOTPSecret} {
		if value.IsUnknown() {
			return
		}
	}

	hasUsername := data.Username.ValueString() != ""
	hasPassword := data.Password.ValueString() != ""

	if hasUsername && !hasPassword {
		diags.AddAttributeError(
			path.Root("password"),
			"password required",
			"password is required when username is provided",
		)
	}

	if hasPassword && !hasUsername {
		diags.AddAttributeError(
			path.Root("username"),
			"username required",
			"username is required when password is provided",
		)
	}

	if data.TOTPSecret.ValueString() != "" && (!hasUsername || !hasPassword) {
		diags.AddAttributeError(
			path.Root("totp_secret"),
			"username and password required",
			"totp_secret is the secret the one-time code of a password login is derived from, so it "+
				"requires username and password. A session_token needs no one-time code, because it "+
				"bypasses two-factor authentication.",
		)
	}
}

// warnRejectedSessionToken reports a configured session_token the server
// refused, which the password login then took over from. rejection is the
// error the token was refused with, or nil for a token the server accepted.
//
// Nothing else says so: the connection succeeds, and the client goes on with a
// fresh token of its own. A configuration whose stored token has died
// therefore keeps applying for as long as the password stays next to it - until
// the day the password is removed and the apply fails instead.
//
// A connection served from the pool does not repeat the warning, because the
// login it would report happened once, on the Configure that opened it.
func warnRejectedSessionToken(rejection error, diags *diag.Diagnostics) {
	if rejection == nil {
		return
	}

	diags.AddAttributeWarning(
		path.Root("session_token"),
		"session token rejected, fell back to password",
		"Uptime Kuma refused the configured session_token and the provider logged in with username "+
			"and password instead. Replace it with the token of a fresh login, or remove the attribute. "+
			"On Uptime Kuma 2.x a token is invalidated by a password change; on Uptime Kuma 3.x it also "+
			"lapses after 7 days without use.\n\n"+
			"Reason given: "+rejection.Error(),
	)
}

// passwordFallbackAdvice returns the summary and the next step for the error
// the client returns when a refused session token was followed by a password
// login that failed too, see client.PasswordFallbackCause. That error wraps
// kuma.ErrInvalidSessionToken whatever the password login failed with, so it
// is classified by its password half: classifying it by the token rejection
// would tell the user to configure the username and password they already
// have, and bury the reason the recovery failed.
//
// It returns empty strings for every other error, including a token rejection
// on its own: the client only falls back when there is a password to offer,
// and never when the token itself was refused with kuma.ErrUserInactive, which
// no password recovers from. The password login can still fail with it, on
// Uptime Kuma 3.x for a banned account whose session lapsed, so the password
// half is matched against it like against every other typed error.
func passwordFallbackAdvice(err error) (summary string, nextStep string) {
	passwordErr := client.PasswordFallbackCause(err)
	if passwordErr == nil {
		return "", ""
	}

	const prefix = "Uptime Kuma refused the configured `session_token`, and the `username` and " +
		"`password` login the provider fell back to failed as well: "

	var rateLimited *kuma.RateLimitedError

	switch {
	case errors.Is(passwordErr, kuma.ErrUserInactive):
		return "Uptime Kuma rejected the session token and the account is inactive", prefix +
			"the account is deactivated or deleted. Reactivate it in Uptime Kuma or configure the " +
			"credentials of an active account."

	case errors.Is(passwordErr, kuma.ErrInvalidCredentials):
		return "Uptime Kuma rejected the session token and the password", prefix +
			"the username or password was rejected too. A password change is what invalidates a " +
			"session token, so check `password` first, then replace `session_token` with the token of " +
			"a fresh login or remove it."

	case errors.Is(passwordErr, kuma.ErrTwoFactorRequired):
		return "Uptime Kuma rejected the session token and requires two-factor authentication", prefix +
			"the account has two-factor authentication enabled and asked for a one-time code. Set " +
			"`totp_secret` if it is not set; if it is, the two-factor step expired before the code was " +
			"sent, so run Terraform again. Replace `session_token` with the token of a fresh login " +
			"either way."

	case errors.Is(passwordErr, kuma.ErrInvalidTOTPCode):
		return "Uptime Kuma rejected the session token and the two-factor authentication code", prefix +
			"the one-time code derived from the TOTP secret was refused. On Uptime Kuma 2.x a code the " +
			"server has already accepted is refused too, and the wait for the next one fails when " +
			"`timeout` or `per_attempt_timeout` is too short, see the attribute documentation of " +
			"`totp_secret`. Otherwise check `totp_secret` and the clock of this host. Replace " +
			"`session_token` with the token of a fresh login either way."

	case errors.Is(passwordErr, kuma.ErrTwoFactorLocked):
		return "Uptime Kuma rejected the session token and locked the two-factor authentication", prefix +
			"the server refuses further two-factor codes for this login after too many wrong ones. Check " +
			"`totp_secret` and the clock of this host, then run Terraform again, and replace " +
			"`session_token` with the token of a fresh login."

	case errors.As(passwordErr, &rateLimited) && rateLimited.RetryAfter > 0:
		return "Uptime Kuma rejected the session token and rate limited the login", prefix +
			fmt.Sprintf("too many logins were attempted from this address. Wait %s before running "+
				"Terraform again, and replace `session_token` with the token of a fresh login either way.",
				rateLimited.RetryAfter)

	case errors.Is(passwordErr, kuma.ErrRateLimited):
		return "Uptime Kuma rejected the session token and rate limited the login", prefix +
			"too many logins were attempted from this address. Wait a minute before running Terraform " +
			"again, and replace `session_token` with the token of a fresh login either way."
	}

	return "Uptime Kuma rejected the session token and the password login failed", prefix +
		"for a reason the provider does not recognize, see the underlying error. That is mostly a " +
		"transport problem: the endpoint URL is incorrect or not reachable from this host, a reverse " +
		"proxy is not forwarding the Socket.IO/WebSocket connection, or a TLS certificate is not " +
		"accepted. Replace `session_token` with the token of a fresh login either way. To collect " +
		"diagnostics, re-run with TF_LOG=DEBUG SOCKETIO_LOG_LEVEL=DEBUG."
}
