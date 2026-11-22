package provider

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// errorSummariesWithPath renders every error as "summary@path", for tests that
// pin which attribute an error names.
func errorSummariesWithPath(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	got := make([]string, 0, diags.ErrorsCount())

	for _, d := range diags.Errors() {
		withPath, ok := d.(diag.DiagnosticWithPath)
		if !ok {
			t.Fatalf("expected error %q to name an attribute", d.Summary())
		}

		got = append(got, d.Summary()+"@"+withPath.Path().String())
	}

	return got
}

// TestValidateAPIKey pins that api_key is the only credential of a provider
// that sets it, and that it rules out the 2.x login.
func TestValidateAPIKey(t *testing.T) {
	t.Parallel()

	key := types.StringValue("uk1_key")

	tests := []struct {
		name       string
		model      func(*UptimeKumaProviderModel)
		wantErrors []string
	}{
		{name: "api key alone", model: func(m *UptimeKumaProviderModel) { m.APIKey = key }},
		{name: "no api key", model: func(m *UptimeKumaProviderModel) { m.Username = types.StringValue("admin") }},
		{
			name: "empty api key is no credential",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = types.StringValue("")
				m.Username = types.StringValue("admin")
			},
		},
		{
			name: "unknown api key is not judged",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = types.StringUnknown()
				m.Username = types.StringValue("admin")
			},
		},
		{
			name: "unknown username is not judged",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = key
				m.Username = types.StringUnknown()
			},
		},
		{
			name: "api key with every other credential",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = key
				m.Username = types.StringValue("admin")
				m.Password = types.StringValue("secret")
				m.SessionToken = types.StringValue("token")
				m.TOTPSecret = types.StringValue("JBSWY3DPEHPK3PXP")
			},
			wantErrors: []string{
				"api_key conflicts with username@username",
				"api_key conflicts with password@password",
				"api_key conflicts with session_token@session_token",
				"api_key conflicts with totp_secret@totp_secret",
			},
		},
		{
			name: "api key with the 2.x login pinned",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = key
				m.AuthScheme = types.StringValue(authSchemeSocket)
			},
			wantErrors: []string{"api_key requires Uptime Kuma 3.0@auth_scheme"},
		},
		{
			name: "api key with an Authorization header",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = key
				m.Headers = types.MapValueMust(types.StringType, map[string]attr.Value{
					"authorization": types.StringValue("Basic dXNlcjpwYXNz"),
					"X-Other":       types.StringValue("x"),
				})
			},
			wantErrors: []string{`api_key conflicts with an Authorization header@headers["authorization"]`},
		},
		{
			name: "api key with the 3.x login pinned",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = key
				m.AuthScheme = types.StringValue(authSchemeHTTP)
			},
		},
		{
			name: "api key with auto detection",
			model: func(m *UptimeKumaProviderModel) {
				m.APIKey = key
				m.AuthScheme = types.StringValue(authSchemeAuto)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			model := nullProviderModel()
			tc.model(&model)

			var diags diag.Diagnostics

			validateAPIKey(&model, nil, &diags)

			got := errorSummariesWithPath(t, diags)
			if strings.Join(got, ",") != strings.Join(tc.wantErrors, ",") {
				t.Errorf("expected errors %v, got %v", tc.wantErrors, got)
			}
		})
	}
}

// TestValidateConfig_APIKey pins what `terraform validate` reports about
// api_key and UPTIMEKUMA_HEADERS.
func TestValidateConfig_APIKey(t *testing.T) {
	str := func(value string) tftypes.Value { return tftypes.NewValue(tftypes.String, value) }

	headers := func(values map[string]tftypes.Value) tftypes.Value {
		return tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, values)
	}

	tests := []struct {
		name      string
		env       map[string]string
		values    map[string]tftypes.Value
		wantError string
	}{
		{
			name:      "api key and username in config",
			values:    map[string]tftypes.Value{"api_key": str("uk1_key"), "username": str("admin")},
			wantError: "api_key conflicts with username",
		},
		{
			name:      "api key and socket scheme in config",
			values:    map[string]tftypes.Value{"api_key": str("uk1_key"), "auth_scheme": str("socket")},
			wantError: "api_key requires Uptime Kuma 3.0",
		},
		{
			// The environment of `terraform validate` need not be the one of
			// the apply; Configure judges the environment.
			name:   "api key in config, password in environment is left to Configure",
			env:    map[string]string{"UPTIMEKUMA_PASSWORD": "secret"},
			values: map[string]tftypes.Value{"api_key": str("uk1_key")},
		},
		{
			name: "api key with an unknown username",
			values: map[string]tftypes.Value{
				"api_key":  str("uk1_key"),
				"username": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			},
		},
		{
			name:      "malformed headers from environment",
			env:       map[string]string{headersEnvironmentVariable: `{"X-Test":`},
			wantError: "invalid UPTIMEKUMA_HEADERS",
		},
		{
			name: "malformed headers from environment shadowed by config",
			env:  map[string]string{headersEnvironmentVariable: `{"X-Test":`},
			values: map[string]tftypes.Value{
				"headers": headers(map[string]tftypes.Value{"X-Test": str("value")}),
			},
		},
		{
			name: "valid headers from environment",
			env:  map[string]string{headersEnvironmentVariable: `{"X-Test":"value"}`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearCredentialEnvironment(t)

			for name, value := range tc.env {
				t.Setenv(name, value)
			}

			req := provider.ValidateConfigRequest{Config: providerConfigValue(t, tc.values)}

			var resp provider.ValidateConfigResponse

			(&UptimeKumaProvider{}).ValidateConfig(t.Context(), req, &resp)

			if tc.wantError == "" {
				if resp.Diagnostics.HasError() {
					t.Errorf("expected no error, got %v", resp.Diagnostics.Errors())
				}

				return
			}

			if count := resp.Diagnostics.ErrorsCount(); count != 1 {
				t.Fatalf("expected exactly one error %q, got %v", tc.wantError, resp.Diagnostics.Errors())
			}

			if got := resp.Diagnostics.Errors()[0].Summary(); got != tc.wantError {
				t.Errorf("expected error %q, got %q", tc.wantError, got)
			}
		})
	}
}

// TestConfigure_APIKeyFailsBeforeConnecting pins that a credential from the
// environment next to api_key is reported by Configure, before it connects.
// The endpoint is a port nothing listens on, see
// TestConfigure_FailsBeforeConnecting.
func TestConfigure_APIKeyFailsBeforeConnecting(t *testing.T) {
	str := func(value string) tftypes.Value { return tftypes.NewValue(tftypes.String, value) }

	tests := []struct {
		name        string
		env         map[string]string
		values      map[string]tftypes.Value
		wantSummary string
		wantPath    path.Path
		wantDetail  string
	}{
		{
			name:        "api key and session token both from the environment",
			env:         map[string]string{"UPTIMEKUMA_API_KEY": "uk1_key", "UPTIMEKUMA_SESSION_TOKEN": "token"},
			wantSummary: "api_key conflicts with session_token",
			wantPath:    path.Root("session_token"),
			wantDetail: "the environment variable UPTIMEKUMA_API_KEY and the environment variable " +
				"UPTIMEKUMA_SESSION_TOKEN are both set",
		},
		{
			name:        "api key and session token both in config",
			values:      map[string]tftypes.Value{"api_key": str("uk1_key"), "session_token": str("token")},
			wantSummary: "api_key conflicts with session_token",
			wantPath:    path.Root("session_token"),
			wantDetail:  "the provider attribute api_key and the provider attribute session_token are both set",
		},
		{
			name:        "authorization header from the environment next to api key",
			env:         map[string]string{headersEnvironmentVariable: `{"Authorization":"Basic eDp5"}`},
			values:      map[string]tftypes.Value{"api_key": str("uk1_key")},
			wantSummary: "api_key conflicts with an Authorization header",
			wantPath:    path.Root("headers").AtMapKey("Authorization"),
			wantDetail:  "the environment variable UPTIMEKUMA_HEADERS cannot set one",
		},
		{
			name:        "user agent in headers from the environment",
			env:         map[string]string{headersEnvironmentVariable: `{"user-agent":"x/1"}`},
			wantSummary: "User-Agent in headers",
			wantPath:    path.Root("headers").AtMapKey("user-agent"),
			wantDetail:  "user_agent",
		},
		{
			name:        "empty header value from the environment",
			env:         map[string]string{headersEnvironmentVariable: `{"X-Token":""}`},
			wantSummary: "empty header value",
			wantPath:    path.Root("headers").AtMapKey("X-Token"),
			wantDetail:  "UPTIMEKUMA_HEADERS",
		},
		{
			name: "unknown headers",
			values: map[string]tftypes.Value{
				"headers": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
			},
			wantSummary: "unknown headers",
			wantPath:    path.Root("headers"),
		},
		{
			name: "unknown header value",
			values: map[string]tftypes.Value{
				"headers": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
					"X-Token": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				}),
			},
			wantSummary: "unknown headers",
			wantPath:    path.Root("headers"),
		},
		{
			name:        "socket scheme from the environment",
			env:         map[string]string{"UPTIMEKUMA_AUTH_SCHEME": "socket"},
			values:      map[string]tftypes.Value{"api_key": str("uk1_key")},
			wantSummary: "api_key requires Uptime Kuma 3.0",
			wantPath:    path.Root("auth_scheme"),
		},
		{
			name:        "malformed headers from the environment",
			env:         map[string]string{headersEnvironmentVariable: `["X-Test"]`},
			values:      map[string]tftypes.Value{"api_key": str("uk1_key")},
			wantSummary: "invalid UPTIMEKUMA_HEADERS",
			wantPath:    path.Root("headers"),
		},
		{
			name:        "unknown api key",
			values:      map[string]tftypes.Value{"api_key": tftypes.NewValue(tftypes.String, tftypes.UnknownValue)},
			wantSummary: "unknown api_key",
			wantPath:    path.Root("api_key"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearCredentialEnvironment(t)
			t.Setenv("UPTIMEKUMA_ENDPOINT", "")

			for name, value := range tc.env {
				t.Setenv(name, value)
			}

			values := map[string]tftypes.Value{"endpoint": str("http://127.0.0.1:1")}
			maps.Copy(values, tc.values)

			req := provider.ConfigureRequest{Config: providerConfigValue(t, values)}

			var resp provider.ConfigureResponse

			(&UptimeKumaProvider{}).Configure(t.Context(), req, &resp)

			want := tc.wantSummary + "@" + tc.wantPath.String()
			if got := errorSummariesWithPath(t, resp.Diagnostics); len(got) != 1 || got[0] != want {
				t.Fatalf("expected exactly the error %q, got %v", want, got)
			}

			if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, tc.wantDetail) {
				t.Errorf("expected the detail to contain %q, got %q", tc.wantDetail, detail)
			}

			if resp.ResourceData != nil {
				t.Error("expected no provider data after a failed Configure")
			}
		})
	}
}

// TestApplyHeadersEnvironmentDefault pins how UPTIMEKUMA_HEADERS is read: as a
// JSON object of strings, where the attribute is null or empty, and never into
// an error message that quotes its values.
func TestApplyHeadersEnvironmentDefault(t *testing.T) {
	configured := types.MapValueMust(types.StringType, map[string]attr.Value{
		"X-From-Config": types.StringValue("config"),
	})

	tests := []struct {
		name        string
		env         string
		headers     types.Map
		want        map[string]string
		wantUnknown bool
		wantError   bool
	}{
		{name: "unset", headers: types.MapNull(types.StringType)},
		{
			name:    "from environment",
			env:     `{"CF-Access-Client-Id":"id.access","CF-Access-Client-Secret":"s3cr3t"}`,
			headers: types.MapNull(types.StringType),
			want:    map[string]string{"CF-Access-Client-Id": "id.access", "CF-Access-Client-Secret": "s3cr3t"},
		},
		{
			name:    "empty map falls back to environment",
			env:     `{"X-Env":"env"}`,
			headers: types.MapValueMust(types.StringType, map[string]attr.Value{}),
			want:    map[string]string{"X-Env": "env"},
		},
		{
			name:    "config takes precedence",
			env:     `{"X-Env":"env"}`,
			headers: configured,
			want:    map[string]string{"X-From-Config": "config"},
		},
		{
			name:        "unknown is left for Configure",
			env:         `{"X-Env":"env"}`,
			headers:     types.MapUnknown(types.StringType),
			wantUnknown: true,
		},
		{name: "invalid JSON", env: `{"X-Secret":"s3cr3t"`, headers: types.MapNull(types.StringType), wantError: true},
		{name: "not an object", env: `["s3cr3t"]`, headers: types.MapNull(types.StringType), wantError: true},
		{name: "non-string value", env: `{"X-Count":12345}`, headers: types.MapNull(types.StringType), wantError: true},
		{name: "null", env: `null`, headers: types.MapNull(types.StringType), wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(headersEnvironmentVariable, tc.env)

			model := nullProviderModel()
			model.Headers = tc.headers

			var diags diag.Diagnostics

			applyHeadersEnvironmentDefault(&model, &diags)

			if diags.HasError() != tc.wantError {
				t.Fatalf("expected an error %t, got %v", tc.wantError, diags.Errors())
			}

			if tc.wantError {
				if detail := diags.Errors()[0].Detail(); strings.Contains(detail, "s3cr3t") ||
					strings.Contains(detail, "12345") {
					t.Errorf("expected the error not to quote the variable's values, got %q", detail)
				}

				return
			}

			if model.Headers.IsUnknown() != tc.wantUnknown {
				t.Fatalf("expected unknown %t, got %v", tc.wantUnknown, model.Headers)
			}

			got := headersFromModel(model.Headers)
			if len(got) != len(tc.want) {
				t.Fatalf("expected headers %v, got %v", tc.want, got)
			}

			for name, value := range tc.want {
				if got[name] != value {
					t.Errorf("expected header %s=%q, got %q", name, value, got[name])
				}
			}
		})
	}
}

func TestApplyEnvironmentDefaults_APIKey(t *testing.T) {
	t.Setenv("UPTIMEKUMA_API_KEY", "uk1_from_env")

	model := nullProviderModel()

	applyEnvironmentDefaults(&model, &diag.Diagnostics{})

	if got := model.APIKey.ValueString(); got != "uk1_from_env" {
		t.Errorf("expected api_key %q from env, got %q", "uk1_from_env", got)
	}

	model = nullProviderModel()
	model.APIKey = types.StringValue("uk1_from_config")

	applyEnvironmentDefaults(&model, &diag.Diagnostics{})

	if got := model.APIKey.ValueString(); got != "uk1_from_config" {
		t.Errorf("expected config api_key %q to take precedence, got %q", "uk1_from_config", got)
	}
}

func TestResolveUserAgent(t *testing.T) {
	t.Parallel()

	model := nullProviderModel()

	if got := resolveUserAgent(&model, "1.2.3"); got != "terraform-provider-uptimekuma/1.2.3" {
		t.Errorf("expected the default user agent, got %q", got)
	}

	model.UserAgent = types.StringValue("")

	if got := resolveUserAgent(&model, "dev"); got != "terraform-provider-uptimekuma/dev" {
		t.Errorf("expected an empty user_agent to get the default, got %q", got)
	}

	model.UserAgent = types.StringValue("my-pipeline/1.0")

	if got := resolveUserAgent(&model, "1.2.3"); got != "my-pipeline/1.0" {
		t.Errorf("expected the configured user agent, got %q", got)
	}
}

// TestBuildClientConfig_RequestHeaders pins that the API key, the headers and
// the User-Agent reach the client configuration.
func TestBuildClientConfig_RequestHeaders(t *testing.T) {
	t.Parallel()

	model := nullProviderModel()
	model.Endpoint = types.StringValue("http://localhost:3001")
	model.APIKey = types.StringValue("uk1_key")
	model.Headers = types.MapValueMust(types.StringType, map[string]attr.Value{
		"CF-Access-Client-Id": types.StringValue("id.access"),
	})

	config := buildClientConfig(&model, clientOptions{userAgent: "terraform-provider-uptimekuma/test"})

	if config.APIKey != "uk1_key" {
		t.Errorf("expected api key %q, got %q", "uk1_key", config.APIKey)
	}

	if len(config.Headers) != 1 || config.Headers["CF-Access-Client-Id"] != "id.access" {
		t.Errorf("expected the configured headers, got %v", config.Headers)
	}

	if config.UserAgent != "terraform-provider-uptimekuma/test" {
		t.Errorf("expected the resolved user agent, got %q", config.UserAgent)
	}

	empty := nullProviderModel()

	if headers := buildClientConfig(&empty, clientOptions{}).Headers; headers != nil {
		t.Errorf("expected no headers for a null map, got %v", headers)
	}
}

func TestRejectUnknownHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers types.Map
		want    bool
	}{
		{name: "null", headers: types.MapNull(types.StringType)},
		{name: "known", headers: types.MapValueMust(types.StringType, map[string]attr.Value{"X": types.StringValue("v")})},
		{name: "unknown map", headers: types.MapUnknown(types.StringType), want: true},
		{
			name:    "unknown value",
			headers: types.MapValueMust(types.StringType, map[string]attr.Value{"X": types.StringUnknown()}),
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			rejectUnknownHeaders(tc.headers, &diags)

			if diags.HasError() != tc.want {
				t.Errorf("expected an error %t, got %v", tc.want, diags.Errors())
			}
		})
	}
}

func TestSettingsPasswordRequiredDetail(t *testing.T) {
	t.Parallel()

	if detail := (&SettingsResource{apiKey: true}).passwordRequiredDetail(); !strings.Contains(detail, "api_key") ||
		!strings.Contains(detail, "alias") {
		t.Errorf("expected the detail to name api_key and a provider alias, got %q", detail)
	}

	if detail := (&SettingsResource{}).passwordRequiredDetail(); strings.Contains(detail, "api_key") ||
		!strings.Contains(detail, "session_token") {
		t.Errorf("expected the detail to name session_token, not api_key, got %q", detail)
	}
}

// TestApplyCredentialEnvironmentDefaults_Groups pins that api_key and the
// login credentials fall back to the environment by group: a configuration
// that sets one group keeps the variables of the other unread, so a provider
// block with api_key ignores a username and password in the environment, and
// an alias with a username and password ignores UPTIMEKUMA_API_KEY.
func TestApplyCredentialEnvironmentDefaults_Groups(t *testing.T) {
	env := map[string]string{
		"UPTIMEKUMA_USERNAME":      "admin-from-env",
		"UPTIMEKUMA_PASSWORD":      "password-from-env",
		"UPTIMEKUMA_SESSION_TOKEN": "token-from-env",
		"UPTIMEKUMA_TOTP_SECRET":   "JBSWY3DPEHPK3PXP",
		"UPTIMEKUMA_API_KEY":       "uk1_from_env",
	}

	tests := []struct {
		name        string
		model       func(*UptimeKumaProviderModel)
		wantLogin   bool
		wantAPIKey  string
		wantSources []string
	}{
		{
			name:        "api key in config ignores the login variables",
			model:       func(m *UptimeKumaProviderModel) { m.APIKey = types.StringValue("uk1_from_config") },
			wantAPIKey:  "uk1_from_config",
			wantSources: nil,
		},
		{
			name:        "unknown api key in config ignores the login variables",
			model:       func(m *UptimeKumaProviderModel) { m.APIKey = types.StringUnknown() },
			wantSources: nil,
		},
		{
			name:        "username in config ignores the api key variable",
			model:       func(m *UptimeKumaProviderModel) { m.Username = types.StringValue("admin") },
			wantLogin:   true,
			wantSources: []string{"password", "session_token", "totp_secret"},
		},
		{
			name:        "session token in config ignores the api key variable",
			model:       func(m *UptimeKumaProviderModel) { m.SessionToken = types.StringValue("token") },
			wantLogin:   true,
			wantSources: []string{"password", "totp_secret", "username"},
		},
		{
			// An empty attribute counts as unset, here as everywhere else.
			name:        "empty api key in config reads both",
			model:       func(m *UptimeKumaProviderModel) { m.APIKey = types.StringValue("") },
			wantLogin:   true,
			wantAPIKey:  "uk1_from_env",
			wantSources: []string{"api_key", "password", "session_token", "totp_secret", "username"},
		},
		{
			name:        "empty username in config reads both",
			model:       func(m *UptimeKumaProviderModel) { m.Username = types.StringValue("") },
			wantLogin:   true,
			wantAPIKey:  "uk1_from_env",
			wantSources: []string{"api_key", "password", "session_token", "totp_secret", "username"},
		},
		{
			name:        "nothing in config reads both, for validateAPIKey to reject",
			model:       func(*UptimeKumaProviderModel) {},
			wantLogin:   true,
			wantAPIKey:  "uk1_from_env",
			wantSources: []string{"api_key", "password", "session_token", "totp_secret", "username"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearCredentialEnvironment(t)

			for name, value := range env {
				t.Setenv(name, value)
			}

			model := nullProviderModel()
			tc.model(&model)

			sources := applyCredentialEnvironmentDefaults(&model)

			if got := model.Password.ValueString() != ""; got != tc.wantLogin {
				t.Errorf("expected a password %t, got %q", tc.wantLogin, model.Password.ValueString())
			}

			if !model.APIKey.IsUnknown() && model.APIKey.ValueString() != tc.wantAPIKey {
				t.Errorf("expected api key %q, got %q", tc.wantAPIKey, model.APIKey.ValueString())
			}

			got := slices.Sorted(maps.Keys(sources))
			if strings.Join(got, ",") != strings.Join(tc.wantSources, ",") {
				t.Errorf("expected sources %v, got %v", tc.wantSources, got)
			}
		})
	}
}

// TestValidateHeaders pins what the provider rejects in `headers` before it
// connects. The headers the client reserves are left to the client.
func TestValidateHeaders(t *testing.T) {
	t.Parallel()

	headers := func(values map[string]attr.Value) types.Map {
		return types.MapValueMust(types.StringType, values)
	}

	tests := []struct {
		name       string
		headers    types.Map
		wantErrors []string
	}{
		{name: "null", headers: types.MapNull(types.StringType)},
		{name: "unknown", headers: types.MapUnknown(types.StringType)},
		{
			name:    "valid",
			headers: headers(map[string]attr.Value{"CF-Access-Client-Id": types.StringValue("id")}),
		},
		{
			// The client owns this list; the provider does not repeat it.
			name:    "reserved name is left to the client",
			headers: headers(map[string]attr.Value{"Host": types.StringValue("example.com")}),
		},
		{
			name:    "unknown value is left for Configure",
			headers: headers(map[string]attr.Value{"X-Token": types.StringUnknown()}),
		},
		{
			name:       "null value",
			headers:    headers(map[string]attr.Value{"X-Token": types.StringNull()}),
			wantErrors: []string{`empty header value@headers["X-Token"]`},
		},
		{
			name:       "empty value",
			headers:    headers(map[string]attr.Value{"X-Token": types.StringValue("")}),
			wantErrors: []string{`empty header value@headers["X-Token"]`},
		},
		{
			name:       "invalid value",
			headers:    headers(map[string]attr.Value{"X-Token": types.StringValue("a\nb")}),
			wantErrors: []string{`invalid header value@headers["X-Token"]`},
		},
		{
			name:       "invalid name",
			headers:    headers(map[string]attr.Value{"X Token": types.StringValue("v")}),
			wantErrors: []string{`invalid header name@headers["X Token"]`},
		},
		{
			name: "names that differ only in case",
			headers: headers(map[string]attr.Value{
				"X-Token": types.StringValue("a"), "x-token": types.StringValue("b"),
			}),
			wantErrors: []string{`duplicate header name@headers["x-token"]`},
		},
		{
			name:       "user agent",
			headers:    headers(map[string]attr.Value{"USER-AGENT": types.StringValue("x/1")}),
			wantErrors: []string{`User-Agent in headers@headers["USER-AGENT"]`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics

			validateHeaders(tc.headers, "", &diags)

			got := errorSummariesWithPath(t, diags)
			if strings.Join(got, ",") != strings.Join(tc.wantErrors, ",") {
				t.Errorf("expected errors %v, got %v", tc.wantErrors, got)
			}
		})
	}
}
