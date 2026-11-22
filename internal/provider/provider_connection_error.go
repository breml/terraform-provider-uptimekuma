package provider

import (
	"errors"
	"fmt"
	"strings"

	kuma "github.com/breml/go-uptime-kuma-client"

	"github.com/breml/terraform-provider-uptimekuma/internal/client"
)

// requestSecrets tells whether the provider sends `headers` or `api_key`,
// credentials of the server or of an access proxy in front of it.
type requestSecrets int

const (
	withoutRequestSecrets requestSecrets = iota
	withRequestSecrets
)

// requestSecretsOf returns whether config sends `headers` or `api_key`.
func requestSecretsOf(config *client.Config) requestSecrets {
	if config.APIKey != "" || len(config.Headers) > 0 {
		return withRequestSecrets
	}

	return withoutRequestSecrets
}

// connectionError returns the summary and detail of the diagnostic for a
// failed connection to Uptime Kuma. A login failure the client reports with a
// typed error gets a diagnostic of its own that names the next step, since
// transport advice does not help with a rejected password. So do a request
// the client refused to make the way it was configured, a refused redirect and
// an auth scheme detection answered by something other than Uptime Kuma, see
// requestConfigErrorAdvice. Every other error, including a login rejection the
// client has no typed error for, gets the generic connection advice, and a
// malformed TOTP secret, which never reached the server, a diagnostic of its
// own.
//
// secrets tells whether the provider sends `headers` or `api_key`, which
// changes what a refused redirect most likely means. The endpoint is
// shown with its password redacted.
func connectionError(endpoint string, err error, secrets requestSecrets) (summary string, detail string) {
	endpoint = client.RedactURL(endpoint)

	// The secret is rejected before any connection is attempted, so neither
	// the login nor the transport advice fits it. The schema validator catches
	// a configured one at plan time; this is the one from the environment.
	if errors.Is(err, kuma.ErrInvalidTOTPSecret) {
		return "Invalid TOTP secret", fmt.Sprintf(
			"The TOTP secret in `totp_secret` (or the environment variable `UPTIMEKUMA_TOTP_SECRET`) is "+
				"not a base32 shared secret, so no one-time code can be derived from it. Copy the `secret` "+
				"out of the `otpauth://` URI Uptime Kuma showed while two-factor authentication was set "+
				"up.\n\nUnderlying error: %v",
			err,
		)
	}

	// The typed errors come first: they are the server's answer, and the
	// message text requestConfigErrorAdvice matches on may occur in them too.
	summary, nextStep := loginErrorAdvice(err)
	if summary != "" {
		return summary, fmt.Sprintf(
			"Uptime Kuma at %q refused the login.\n\n"+
				"Underlying error: %v\n\n"+
				"%s",
			endpoint,
			err,
			nextStep,
		)
	}

	summary, detail = requestConfigErrorAdvice(endpoint, err, secrets)
	if summary != "" {
		return summary, detail
	}

	return "failed to connect to Uptime Kuma", transportErrorDetail(endpoint, err)
}

// apiKeyErrorAdvice returns the summary and the next step for a refused
// `api_key`, or empty strings for any other error. It is the tail of
// loginErrorAdvice.
func apiKeyErrorAdvice(err error) (summary string, nextStep string) {
	switch {
	case errors.Is(err, kuma.ErrAPIKeyUnsupported):
		return "Uptime Kuma does not accept API keys for this connection",
			"API keys authenticate the socket only on Uptime Kuma 3.0, and this server is Uptime Kuma 2.x " +
				"or earlier. Authenticate with `username` and `password` or `session_token` instead of " +
				"`api_key`, or upgrade Uptime Kuma."

	case errors.Is(err, kuma.ErrAPIKeyRejected):
		return "Uptime Kuma rejected the API key",
			"The server asked for a login instead of accepting `api_key` (or `UPTIMEKUMA_API_KEY`): the key " +
				"is invalid, expired or revoked, or \"API and MCP\" is disabled in Settings (`apiEnabled`), " +
				"or `auth_scheme = \"http\"` is pinned against an Uptime Kuma 2.x server, which asks for a " +
				"login the same way. The server does not tell these apart. Check the key under Settings > " +
				"API Keys, that \"API and MCP\" is enabled, and that the server runs Uptime Kuma 3.0."

	case errors.Is(err, kuma.ErrAPIKeySetupRequired):
		return "Uptime Kuma is not set up yet",
			"The server has not been set up, so it has no API keys, and an API key cannot run the setup. " +
				"Set Uptime Kuma up with a username and password first, create an API key, then use it in " +
				"`api_key`."
	}

	return "", ""
}

// requestConfigErrorAdvice returns the summary and the detail of the
// diagnostic for a connection the client refused to make the way it was
// configured, or that something other than Uptime Kuma answered: an endpoint
// with userinfo next to an API key, an `api_key`, `headers` or `user_agent` the
// client rejects, a redirect that leaves the origin of `endpoint`, and an auth
// scheme detection answered by neither Uptime Kuma 2.x nor 3.x. None of them is
// a refused login nor a transport failure. It returns empty strings for any
// other error. endpoint is already redacted.
func requestConfigErrorAdvice(endpoint string, err error, secrets requestSecrets) (summary string, detail string) {
	switch {
	case client.IsUserinfoError(err):
		return "endpoint contains a user name and password", fmt.Sprintf(
			"The endpoint %q carries a user name and password (user:password@), from which the HTTP client "+
				"builds an Authorization header of its own. That cannot be combined with `api_key` or an "+
				"`Authorization` header in `headers`, because the server would see two credentials.\n\n"+
				"Underlying error: %v\n\n"+
				"Remove the user:password@ part from `endpoint` (or `UPTIMEKUMA_ENDPOINT`).",
			endpoint, err,
		)

	case client.IsRequestConfigError(err):
		return "Invalid api_key, headers or user_agent", requestConfigErrorDetail(endpoint, err)

	case client.IsRefusedRedirect(err):
		return redirectErrorAdvice(endpoint, err, secrets)

	case client.IsAuthSchemeProbeError(err):
		return "Unexpected answer to the auth scheme detection", fmt.Sprintf(
			"Before it connects, the provider asks %q whether it runs Uptime Kuma 2.x or 3.x, and the answer "+
				"was neither.\n\n"+
				"Underlying error: %v\n\n"+
				"An access proxy in front of Uptime Kuma may be redirecting the request to its sign-in, or "+
				"rejecting it: check the credentials the proxy expects in `headers` (or "+
				"`UPTIMEKUMA_HEADERS`), such as the service token of Cloudflare Access. A reverse proxy that "+
				"answers this request itself can be bypassed by pinning `auth_scheme` (`http` for Uptime Kuma "+
				"3.x, `socket` for 2.x).",
			endpoint, err,
		)
	}

	return "", ""
}

// requestConfigErrorDetail returns the detail of the diagnostic for an
// `api_key`, `headers` or `user_agent` the client rejects, see
// requestConfigErrorAdvice.
func requestConfigErrorDetail(endpoint string, err error) string {
	detail := fmt.Sprintf(
		"The provider could not connect to Uptime Kuma at %q with the configured `api_key`, `headers` "+
			"or `user_agent` (or `UPTIMEKUMA_API_KEY` and `UPTIMEKUMA_HEADERS`).\n\n"+
			"Underlying error: %v\n\n"+
			"`headers` cannot set what the client and its transport set themselves (Host, Origin, "+
			"Connection, Upgrade, Accept, Accept-Encoding, Content-Type, Content-Length, "+
			"Transfer-Encoding, TE, Trailer and Sec-WebSocket-*), nor an `Authorization: Bearer` token, "+
			"which belongs in `api_key`. Names and values must be valid HTTP header fields. `api_key` is "+
			"passed without the `Bearer ` prefix.",
		endpoint, err,
	)

	if strings.Contains(err.Error(), "whitespace or control characters") {
		detail += " The key contains whitespace or a control character, most often the line break at " +
			"the end of a file: use trimspace() when reading the key from a file, as in " +
			"`trimspace(file(\"api-key.txt\"))`."
	}

	return detail
}

// redirectErrorAdvice returns the summary and the detail of the diagnostic
// for a redirect the client refused to follow, see requestConfigErrorAdvice.
//
// An upgrade from http to https is a wrong scheme in the endpoint. A redirect
// to another host is too, unless the provider sends `headers` or `api_key`:
// then it is most likely an access proxy that rejected the request and sends
// it to its sign-in page, and the endpoint is right.
func redirectErrorAdvice(endpoint string, err error, secrets requestSecrets) (summary string, detail string) {
	const postsReason = "A redirect would turn the POST requests of the login and the long polling into GET " +
		"requests without a body, so the provider follows none that leaves the origin (scheme, host and " +
		"port) of `endpoint`."

	reason := postsReason
	if secrets == withRequestSecrets {
		reason += " It also keeps the API key and the configured headers from reaching another origin."
	}

	if client.IsHTTPSUpgradeRedirect(err) {
		return "Uptime Kuma redirected to https", fmt.Sprintf(
			"Uptime Kuma at %q answered with a redirect from http to https.\n\n"+
				"Underlying error: %v\n\n"+
				"%s Set `endpoint` (or `UPTIMEKUMA_ENDPOINT`) to the https URL, as named above.",
			endpoint, err, reason,
		)
	}

	if secrets == withRequestSecrets {
		return "Uptime Kuma redirected to another origin", fmt.Sprintf(
			"The request to %q was redirected to another origin.\n\n"+
				"Underlying error: %v\n\n"+
				"A redirect to another host, such as <team>.cloudflareaccess.com, usually means that an "+
				"access proxy in front of Uptime Kuma rejected the request and sent it to its sign-in: check "+
				"the service token in `headers` (or `UPTIMEKUMA_HEADERS`). Otherwise set `endpoint` (or "+
				"`UPTIMEKUMA_ENDPOINT`) to the URL the redirect leads to, as named above. %s",
			endpoint, err, reason,
		)
	}

	return "Uptime Kuma redirected to another origin", fmt.Sprintf(
		"The request to %q was redirected to another origin.\n\n"+
			"Underlying error: %v\n\n"+
			"%s Set `endpoint` (or `UPTIMEKUMA_ENDPOINT`) to the URL the redirect leads to, as named above.",
		endpoint, err, postsReason,
	)
}
