package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// headersEnvironmentVariable holds the `headers` as a JSON object of strings.
const headersEnvironmentVariable = "UPTIMEKUMA_HEADERS"

// apiKeyDescription documents the `api_key` attribute.
const apiKeyDescription = "API key of an Uptime Kuma 3.0 server, used instead of a login. The provider " +
	"presents it as `Authorization: Bearer <key>` on the Socket.IO handshake, which the server accepts " +
	"only while \"API and MCP\" is enabled in its settings (`apiEnabled`, alpha upstream). It is the only " +
	"credential of a provider that sets it, so it conflicts with `username`, `password`, `session_token` " +
	"and `totp_secret` (including their environment variables), and with `auth_scheme = \"socket\"`: " +
	"Uptime Kuma 2.x has no API key login on the Socket.IO API. A server with authentication disabled " +
	"ignores the key. `uptimekuma_settings` cannot be written with it, because Uptime Kuma asks for the " +
	"account password whenever settings are written. Can be set via `UPTIMEKUMA_API_KEY` environment " +
	"variable, which applies when the attribute is unset or empty."

// headersDescription documents the `headers` attribute.
const headersDescription = "HTTP headers added to every request to Uptime Kuma, the Socket.IO handshake " +
	"included, for an access proxy in front of it such as Cloudflare Access (`CF-Access-Client-Id` and " +
	"`CF-Access-Client-Secret`). They are sent to the origin (scheme, host and port) of `endpoint` only, " +
	"and the provider follows no redirect that leaves that origin, an upgrade from `http` to `https` " +
	"included: set `endpoint` to the URL the redirect names instead. The headers the client and its " +
	"transport set themselves (`Host`, `Origin`, `Connection`, `Upgrade`, `Accept`, `Accept-Encoding`, " +
	"`Content-Type`, `Content-Length`, `Transfer-Encoding`, `TE`, `Trailer` and `Sec-WebSocket-*`) are " +
	"rejected, and so is an `Authorization` header with a `Bearer` token, which belongs in `api_key`; " +
	"`Authorization: Basic` for a reverse proxy is accepted. A `User-Agent` here is overridden by " +
	"`user_agent`. Can be set via `UPTIMEKUMA_HEADERS` environment variable as a JSON object of strings " +
	"(e.g. `{\"CF-Access-Client-Id\":\"...\"}`), which applies when the attribute is unset or empty."

// userAgentDescription documents the `user_agent` attribute.
const userAgentDescription = "User-Agent of every request to Uptime Kuma. Defaults to " +
	"`terraform-provider-uptimekuma/<provider version>`."

// defaultUserAgent is the User-Agent of a provider without `user_agent`.
func defaultUserAgent(version string) string {
	return "terraform-provider-uptimekuma/" + version
}

// resolveUserAgent returns `user_agent`, or the default for the provider
// version where it is unset or empty. Provider schemas have no defaults, so it
// is applied here, in Configure.
func resolveUserAgent(data *UptimeKumaProviderModel, version string) string {
	if data.UserAgent.ValueString() != "" {
		return data.UserAgent.ValueString()
	}

	return defaultUserAgent(version)
}

// applyHeadersEnvironmentDefault takes `headers` from UPTIMEKUMA_HEADERS where
// the attribute is null or an empty map, the way the credentials fall back to
// the environment when null or empty, see stringFromEnvironment. An unknown
// value is left alone; Configure rejects it. A value that is not a JSON object
// of strings is reported as an error on `headers`.
func applyHeadersEnvironmentDefault(data *UptimeKumaProviderModel, diags *diag.Diagnostics) {
	if data.Headers.IsUnknown() || len(data.Headers.Elements()) > 0 {
		return
	}

	env := os.Getenv(headersEnvironmentVariable)
	if env == "" {
		return
	}

	var headers map[string]string

	err := json.Unmarshal([]byte(env), &headers)
	if err != nil || headers == nil {
		if err == nil {
			err = errors.New("got a JSON null")
		}

		diags.AddAttributeError(
			path.Root("headers"),
			"invalid "+headersEnvironmentVariable,
			fmt.Sprintf(
				"The environment variable %s must be a JSON object whose values are strings, such as "+
					`{"CF-Access-Client-Id":"id.access","CF-Access-Client-Secret":"secret"}.`+
					"\n\nUnderlying error: %v",
				headersEnvironmentVariable, sanitizeJSONError(err),
			),
		)

		return
	}

	elements := make(map[string]attr.Value, len(headers))
	for name, value := range headers {
		elements[name] = types.StringValue(value)
	}

	data.Headers = types.MapValueMust(types.StringType, elements)
}

// sanitizeJSONError keeps a JSON decoding error from quoting the variable's
// content, which may hold secrets.
func sanitizeJSONError(err error) error {
	if syntaxErr, ok := errors.AsType[*json.SyntaxError](err); ok {
		return fmt.Errorf("invalid JSON at offset %d", syntaxErr.Offset)
	}

	if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		if typeErr.Field != "" {
			return fmt.Errorf("the value of %q is a JSON %s, not a string", typeErr.Field, typeErr.Value)
		}

		return fmt.Errorf("the value is a JSON %s, not an object of strings", typeErr.Value)
	}

	return err
}

// headersFromModel returns the configured `headers` as the client takes them.
// A null, unknown or empty map is no headers.
func headersFromModel(headers types.Map) map[string]string {
	if headers.IsNull() || headers.IsUnknown() || len(headers.Elements()) == 0 {
		return nil
	}

	result := make(map[string]string, len(headers.Elements()))

	for name, value := range headers.Elements() {
		str, ok := value.(types.String)
		if ok {
			result[name] = str.ValueString()
		}
	}

	return result
}

// rejectUnknownHeaders reports `headers` where the map or one of its values is
// still unknown when the provider is configured, for the reason
// rejectUnknownConnectionAttributes gives.
func rejectUnknownHeaders(headers types.Map, diags *diag.Diagnostics) {
	unknown := headers.IsUnknown()

	for _, value := range headers.Elements() {
		if value.IsUnknown() {
			unknown = true
		}
	}

	if !unknown {
		return
	}

	diags.AddAttributeError(
		path.Root("headers"),
		"unknown headers",
		"The provider attribute headers refers to a value that is not known until apply, but the provider "+
			"connects to Uptime Kuma during plan, so the headers it connects with must be known at plan "+
			"time. Set headers from a variable, a data source or an environment variable instead of from a "+
			"resource that is created in the same run.",
	)
}

// validateAPIKey rejects what cannot go together with `api_key`: the other
// credentials, which the server would ignore next to the key, and a pinned
// `auth_scheme = "socket"`, the login of Uptime Kuma 2.x, which has no API key
// login on the Socket.IO API. The client rejects both itself, but from inside
// the connection, with an error that reads like a connection failure.
//
// It judges the values it is handed: ValidateConfig passes the configuration
// alone, so `terraform validate` reports a conflict written in the
// configuration, and Configure passes the values with the environment applied,
// so a credential exported next to the key is reported too. An unknown value
// cannot be judged and is skipped, as is an empty one, which is no credential.
func validateAPIKey(data *UptimeKumaProviderModel, diags *diag.Diagnostics) {
	if data.APIKey.IsUnknown() || data.APIKey.ValueString() == "" {
		return
	}

	for _, credential := range []struct {
		name  string
		env   string
		value types.String
	}{
		{name: "username", env: "UPTIMEKUMA_USERNAME", value: data.Username},
		{name: "password", env: "UPTIMEKUMA_PASSWORD", value: data.Password},
		{name: "session_token", env: "UPTIMEKUMA_SESSION_TOKEN", value: data.SessionToken},
		{name: "totp_secret", env: "UPTIMEKUMA_TOTP_SECRET", value: data.TOTPSecret},
	} {
		if credential.value.IsUnknown() || credential.value.ValueString() == "" {
			continue
		}

		diags.AddAttributeError(
			path.Root(credential.name),
			"api_key conflicts with "+credential.name,
			fmt.Sprintf(
				"api_key is the only credential of a provider that sets it: Uptime Kuma authenticates the "+
					"connection by the key and would ignore %[1]s. Remove %[1]s (and the environment "+
					"variable %[2]s), or remove api_key (and UPTIMEKUMA_API_KEY).",
				credential.name, credential.env,
			),
		)
	}

	if !data.AuthScheme.IsUnknown() && data.AuthScheme.ValueString() == authSchemeSocket {
		diags.AddAttributeError(
			path.Root("auth_scheme"),
			"api_key requires Uptime Kuma 3.0",
			"auth_scheme = \"socket\" is the login of Uptime Kuma 2.x, which has no API key login on the "+
				"Socket.IO API. API keys authenticate the socket only on Uptime Kuma 3.0: set auth_scheme to "+
				"\"http\" or leave it unset for the detection.",
		)
	}
}
