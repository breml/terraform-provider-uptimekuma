package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/net/http/httpguts"
)

// headersEnvironmentVariable holds the `headers` as a JSON object of strings.
const headersEnvironmentVariable = "UPTIMEKUMA_HEADERS"

// apiKeyDescription documents the `api_key` attribute.
const apiKeyDescription = "API key of an Uptime Kuma 3.0 server, used instead of a login. The provider " +
	"presents it as `Authorization: Bearer <key>` on the Socket.IO handshake, which the server accepts " +
	"only while \"API and MCP\" is enabled in its settings (`apiEnabled`, alpha upstream). It is the only " +
	"credential of a provider that sets it, so it conflicts with `username`, `password`, `session_token` " +
	"and `totp_secret`, with an `Authorization` header in `headers`, and with `auth_scheme = \"socket\"`: " +
	"Uptime Kuma 2.x has no API key login on the Socket.IO API. Set in the configuration, it keeps the " +
	"provider from reading `UPTIMEKUMA_USERNAME`, `UPTIMEKUMA_PASSWORD`, `UPTIMEKUMA_SESSION_TOKEN` and " +
	"`UPTIMEKUMA_TOTP_SECRET`; a login credential set in the configuration keeps it from reading " +
	"`UPTIMEKUMA_API_KEY`. A server with authentication disabled ignores the key. `uptimekuma_settings` " +
	"cannot be written with it, because Uptime Kuma asks for the account password whenever settings are " +
	"written. Can be set via `UPTIMEKUMA_API_KEY` environment variable, which applies when the attribute " +
	"is unset or empty."

// headersDescription documents the `headers` attribute.
const headersDescription = "HTTP headers added to every request to Uptime Kuma, the Socket.IO handshake " +
	"included, for an access proxy in front of it such as Cloudflare Access (`CF-Access-Client-Id` and " +
	"`CF-Access-Client-Secret`). They are sent to the origin (scheme, host and port) of `endpoint` only, " +
	"and the provider follows no redirect that leaves that origin, an upgrade from `http` to `https` " +
	"included: set `endpoint` to the URL the redirect names instead. The headers the client and its " +
	"transport set themselves (`Host`, `Origin`, `Connection`, `Upgrade`, `Accept`, `Accept-Encoding`, " +
	"`Content-Type`, `Content-Length`, `Transfer-Encoding`, `TE`, `Trailer` and `Sec-WebSocket-*`) are " +
	"rejected, and so is an `Authorization` header with a `Bearer` token, which belongs in `api_key`; " +
	"`Authorization: Basic` for a reverse proxy is accepted, except next to `api_key`. `User-Agent` is " +
	"rejected too, set it with `user_agent`. A value must not be empty, and two names that differ only " +
	"in case are rejected. Can be set via `UPTIMEKUMA_HEADERS` environment variable as a JSON object of strings " +
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
// of strings is reported as an error on `headers`. It reports whether it took
// the headers from the environment.
func applyHeadersEnvironmentDefault(data *UptimeKumaProviderModel, diags *diag.Diagnostics) bool {
	if data.Headers.IsUnknown() || len(data.Headers.Elements()) > 0 {
		return false
	}

	env := os.Getenv(headersEnvironmentVariable)
	if env == "" {
		return false
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

		return false
	}

	elements := make(map[string]attr.Value, len(headers))
	for name, value := range headers {
		elements[name] = types.StringValue(value)
	}

	data.Headers = types.MapValueMust(types.StringType, elements)

	return true
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

// validateAPIKey rejects what cannot go together with `api_key`: the login
// credentials, which the server would ignore next to the key, an
// `Authorization` header, which the key sets itself, and a pinned
// `auth_scheme = "socket"`, the login of Uptime Kuma 2.x, which has no API key
// login on the Socket.IO API. The client rejects all of them itself, but from
// inside the connection, with an error that reads like a connection failure.
//
// It judges the values it is handed. ValidateConfig passes the configuration
// alone (sources nil), so `terraform validate` reports a conflict written in
// the configuration. Configure passes the values with the environment applied,
// and where each came from: applyCredentialEnvironmentDefaults reads the
// environment by group, so a key and a login credential that both reach this
// come from the same source, and the diagnostic names it. An unknown value
// cannot be judged and is skipped, as is an empty one, which is no credential.
func validateAPIKey(data *UptimeKumaProviderModel, sources envSources, diags *diag.Diagnostics) {
	if data.APIKey.IsUnknown() || data.APIKey.ValueString() == "" {
		return
	}

	apiKey := sources.source("api_key")

	for _, credential := range loginCredentialVariables {
		value := *loginCredential(data, credential.attribute)
		if value.IsUnknown() || value.ValueString() == "" {
			continue
		}

		diags.AddAttributeError(
			path.Root(credential.attribute),
			"api_key conflicts with "+credential.attribute,
			fmt.Sprintf(
				"%s and %s are both set, but api_key is the only credential of a provider that sets it: "+
					"Uptime Kuma authenticates the connection by the key and would ignore %s. Remove one "+
					"of them. A credential set in the provider configuration keeps the provider from reading "+
					"the environment variables of the other kind, so a provider block may set api_key while "+
					"the environment holds a username and password, and the other way round.",
				apiKey, sources.source(credential.attribute), credential.attribute,
			),
		)
	}

	for _, name := range sortedHeaderNames(data.Headers) {
		if !strings.EqualFold(name, "Authorization") {
			continue
		}

		diags.AddAttributeError(
			path.Root("headers").AtMapKey(name),
			"api_key conflicts with an Authorization header",
			fmt.Sprintf(
				"%s sets the Authorization header (Bearer) itself, so %s cannot set one too. Remove the "+
					"Authorization header, or authenticate the reverse proxy in front of Uptime Kuma by "+
					"another header.",
				apiKey, sources.source("headers"),
			),
		)
	}

	if !data.AuthScheme.IsUnknown() && data.AuthScheme.ValueString() == authSchemeSocket {
		diags.AddAttributeError(
			path.Root("auth_scheme"),
			"api_key requires Uptime Kuma 3.0",
			fmt.Sprintf(
				"%s pins auth_scheme = \"socket\", the login of Uptime Kuma 2.x, which has no API key login "+
					"on the Socket.IO API, but %s is set. API keys authenticate the socket only on Uptime Kuma "+
					"3.0: set auth_scheme to \"http\" or leave it unset for the detection.",
				sources.source("auth_scheme"), apiKey,
			),
		)
	}
}

// sortedHeaderNames returns the names of `headers`, sorted, so that the
// diagnostics about them come in a stable order.
func sortedHeaderNames(headers types.Map) []string {
	return slices.Sorted(maps.Keys(headers.Elements()))
}

// validateHeaders rejects the `headers` the provider can judge before it
// connects: a name or value that is not a valid HTTP header field, a null or
// empty value, two names that differ only in case, which HTTP treats as one,
// and a User-Agent, which `user_agent` sets. The headers the client and its
// transport set themselves are left to the client, which is the one source of
// that list; its error is reported at connect time, see
// requestConfigErrorAdvice.
//
// env names the environment variable the headers came from, or is empty for
// headers from the configuration. ValidateConfig calls it for the
// configuration, Configure for the headers it connects with. An unknown map or
// value is skipped; Configure rejects it.
func validateHeaders(headers types.Map, env string, diags *diag.Diagnostics) {
	if headers.IsNull() || headers.IsUnknown() {
		return
	}

	source := "headers"
	if env != "" {
		source = "the environment variable " + env
	}

	seen := map[string]string{}

	for _, name := range sortedHeaderNames(headers) {
		at := path.Root("headers").AtMapKey(name)

		if !httpguts.ValidHeaderFieldName(name) {
			diags.AddAttributeError(at, "invalid header name",
				fmt.Sprintf("%q in %s is not a valid HTTP header name.", name, source))

			continue
		}

		canonical := strings.ToLower(name)
		if other, ok := seen[canonical]; ok {
			diags.AddAttributeError(at, "duplicate header name",
				fmt.Sprintf("%s sets %q and %q, which HTTP treats as one header. Keep one of them.",
					source, other, name))
		}

		seen[canonical] = name

		if canonical == "user-agent" {
			diags.AddAttributeError(at, "User-Agent in headers",
				fmt.Sprintf("%s sets the User-Agent header. Set it with the provider attribute user_agent "+
					"instead, which defaults to terraform-provider-uptimekuma/<provider version>.", source))

			continue
		}

		validateHeaderValue(headers.Elements()[name], name, source, at, diags)
	}
}

// validateHeaderValue rejects a null, empty or invalid value of the header
// name, see validateHeaders. The value itself is never quoted, because it is
// often a credential.
func validateHeaderValue(value attr.Value, name string, source string, at path.Path, diags *diag.Diagnostics) {
	if value.IsUnknown() {
		return
	}

	str, ok := value.(types.String)
	if !ok || str.IsNull() || str.ValueString() == "" {
		diags.AddAttributeError(at, "empty header value",
			fmt.Sprintf("The header %q in %s has no value. Give it one, or remove it.", name, source))

		return
	}

	if !httpguts.ValidHeaderFieldValue(str.ValueString()) {
		diags.AddAttributeError(at, "invalid header value",
			fmt.Sprintf("The value of the header %q in %s is not a valid HTTP header value: it contains a "+
				"line break or another control character.", name, source))
	}
}
