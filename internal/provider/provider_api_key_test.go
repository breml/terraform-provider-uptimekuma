package provider

import (
	"maps"
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

			validateAPIKey(&model, &diags)

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
	}{
		{
			name:        "session token from the environment",
			env:         map[string]string{"UPTIMEKUMA_SESSION_TOKEN": "token"},
			values:      map[string]tftypes.Value{"api_key": str("uk1_key")},
			wantSummary: "api_key conflicts with session_token",
			wantPath:    path.Root("session_token"),
		},
		{
			name:        "api key from the environment, session token in config",
			env:         map[string]string{"UPTIMEKUMA_API_KEY": "uk1_key"},
			values:      map[string]tftypes.Value{"session_token": str("token")},
			wantSummary: "api_key conflicts with session_token",
			wantPath:    path.Root("session_token"),
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
