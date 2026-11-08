package provider

import (
	"errors"
	"maps"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// clearCredentialEnvironment unsets every credential environment variable for
// the duration of the test, so a value exported in the shell running the tests
// cannot change the shape under test.
func clearCredentialEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		"UPTIMEKUMA_USERNAME",
		"UPTIMEKUMA_PASSWORD",
		"UPTIMEKUMA_SESSION_TOKEN",
		"UPTIMEKUMA_TOTP_SECRET",
		"UPTIMEKUMA_AUTH_SCHEME",
	} {
		t.Setenv(name, "")
	}
}

// nullProviderModel returns a provider model with every attribute null, which
// is what an empty provider block decodes into.
func nullProviderModel() UptimeKumaProviderModel {
	return UptimeKumaProviderModel{
		Endpoint:          types.StringNull(),
		Username:          types.StringNull(),
		Password:          types.StringNull(),
		SessionToken:      types.StringNull(),
		TOTPSecret:        types.StringNull(),
		AuthScheme:        types.StringNull(),
		Timeout:           types.StringNull(),
		PerAttemptTimeout: types.StringNull(),
		OperationTimeout:  types.StringNull(),
		MaxRetries:        types.Int64Null(),
	}
}

func TestApplyEnvironmentDefaults_Credentials(t *testing.T) {
	t.Setenv("UPTIMEKUMA_SESSION_TOKEN", "token-from-env")
	t.Setenv("UPTIMEKUMA_TOTP_SECRET", "JBSWY3DPEHPK3PXP")
	t.Setenv("UPTIMEKUMA_AUTH_SCHEME", "http")

	model := nullProviderModel()

	applyEnvironmentDefaults(&model, &diag.Diagnostics{})

	if got := model.SessionToken.ValueString(); got != "token-from-env" {
		t.Errorf("expected session_token %q from env, got %q", "token-from-env", got)
	}

	if got := model.TOTPSecret.ValueString(); got != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected totp_secret %q from env, got %q", "JBSWY3DPEHPK3PXP", got)
	}

	if got := model.AuthScheme.ValueString(); got != "http" {
		t.Errorf("expected auth_scheme %q from env, got %q", "http", got)
	}
}

func TestApplyEnvironmentDefaults_CredentialsConfigOverridesEnv(t *testing.T) {
	t.Setenv("UPTIMEKUMA_SESSION_TOKEN", "token-from-env")
	t.Setenv("UPTIMEKUMA_TOTP_SECRET", "JBSWY3DPEHPK3PXP")
	t.Setenv("UPTIMEKUMA_AUTH_SCHEME", "http")

	model := nullProviderModel()
	model.SessionToken = types.StringValue("token-from-config")
	model.TOTPSecret = types.StringValue("MFRGGZDFMZTWQ2LK")
	model.AuthScheme = types.StringValue("socket")

	applyEnvironmentDefaults(&model, &diag.Diagnostics{})

	if got := model.SessionToken.ValueString(); got != "token-from-config" {
		t.Errorf("expected config session_token %q to take precedence, got %q", "token-from-config", got)
	}

	if got := model.TOTPSecret.ValueString(); got != "MFRGGZDFMZTWQ2LK" {
		t.Errorf("expected config totp_secret %q to take precedence, got %q", "MFRGGZDFMZTWQ2LK", got)
	}

	if got := model.AuthScheme.ValueString(); got != "socket" {
		t.Errorf("expected config auth_scheme %q to take precedence, got %q", "socket", got)
	}
}

// TestApplyEnvironmentDefaults_EmptySecretsFallBackToEnv pins that a secret
// attribute fed from an unset variable, which leaves it empty rather than null,
// does not shadow the environment.
func TestApplyEnvironmentDefaults_EmptySecretsFallBackToEnv(t *testing.T) {
	t.Setenv("UPTIMEKUMA_SESSION_TOKEN", "token-from-env")
	t.Setenv("UPTIMEKUMA_TOTP_SECRET", "JBSWY3DPEHPK3PXP")

	model := nullProviderModel()
	model.SessionToken = types.StringValue("")
	model.TOTPSecret = types.StringValue("")

	applyEnvironmentDefaults(&model, &diag.Diagnostics{})

	if got := model.SessionToken.ValueString(); got != "token-from-env" {
		t.Errorf("expected session_token %q from env, got %q", "token-from-env", got)
	}

	if got := model.TOTPSecret.ValueString(); got != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected totp_secret %q from env, got %q", "JBSWY3DPEHPK3PXP", got)
	}
}

// TestValidateCredentials pins the shapes Uptime Kuma can be authenticated
// with. A session token next to a username and password is deliberately
// allowed: the client tries the token first and falls back to the password
// login for a token the server refuses, which is what recovers a configuration
// whose stored token went stale.
func TestValidateCredentials(t *testing.T) {
	t.Parallel()

	null := types.StringNull()
	unknown := types.StringUnknown()
	admin := types.StringValue("admin")
	secret := types.StringValue("secret")
	token := types.StringValue("token")
	totp := types.StringValue("JBSWY3DPEHPK3PXP")
	empty := types.StringValue("")

	tests := []struct {
		name         string
		username     types.String
		password     types.String
		sessionToken types.String
		totpSecret   types.String
		wantErrors   []string
	}{
		{name: "no credentials", username: null, password: null, sessionToken: null, totpSecret: null},
		{name: "username and password", username: admin, password: secret, sessionToken: null, totpSecret: null},
		{
			name:         "username, password and totp secret",
			username:     admin,
			password:     secret,
			sessionToken: null,
			totpSecret:   totp,
		},
		{name: "session token alone", username: null, password: null, sessionToken: token, totpSecret: null},
		{
			name:         "session token with username and password",
			username:     admin,
			password:     secret,
			sessionToken: token,
			totpSecret:   null,
		},
		{name: "all four", username: admin, password: secret, sessionToken: token, totpSecret: totp},
		{
			// An attribute fed from an unset variable is not null, but the
			// client is handed nothing, so it has to count as unset.
			name: "empty totp secret with a session token", username: null, password: null,
			sessionToken: token, totpSecret: empty,
		},
		{
			name:         "empty session token is no credential",
			username:     null,
			password:     null,
			sessionToken: empty,
			totpSecret:   null,
		},
		{
			// An unknown credential cannot be judged yet; Configure judges it
			// once it is known.
			name: "unknown password is not judged", username: admin, password: unknown,
			sessionToken: null, totpSecret: null,
		},
		{
			name: "unknown totp secret is not judged", username: null, password: null,
			sessionToken: token, totpSecret: unknown,
		},
		{
			// The client rejects this pairing itself, with an error that
			// carries no sentinel and arrives looking like a connection
			// failure, so it is worth catching here.
			name: "empty username with a password", username: empty, password: secret,
			sessionToken: null, totpSecret: null, wantErrors: []string{"username required@username"},
		},
		{
			name: "username without password", username: admin, password: null,
			sessionToken: null, totpSecret: null, wantErrors: []string{"password required@password"},
		},
		{
			name: "username and session token without password", username: admin, password: null,
			sessionToken: token, totpSecret: null, wantErrors: []string{"password required@password"},
		},
		{
			name: "password without username", username: null, password: secret,
			sessionToken: null, totpSecret: null, wantErrors: []string{"username required@username"},
		},
		{
			name: "username and totp secret without password", username: admin, password: null,
			sessionToken: null, totpSecret: totp,
			wantErrors: []string{"password required@password", "username and password required@totp_secret"},
		},
		{
			name: "totp secret alone", username: null, password: null,
			sessionToken: null, totpSecret: totp, wantErrors: []string{"username and password required@totp_secret"},
		},
		{
			name: "totp secret with session token only", username: null, password: null,
			sessionToken: token, totpSecret: totp, wantErrors: []string{"username and password required@totp_secret"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			model := nullProviderModel()
			model.Username = tc.username
			model.Password = tc.password
			model.SessionToken = tc.sessionToken
			model.TOTPSecret = tc.totpSecret

			var diags diag.Diagnostics

			validateCredentials(&model, &diags)

			got := make([]string, 0, len(diags.Errors()))
			for _, d := range diags.Errors() {
				withPath, ok := d.(diag.DiagnosticWithPath)
				if !ok {
					t.Fatalf("expected error %q to name an attribute", d.Summary())
				}

				got = append(got, d.Summary()+"@"+withPath.Path().String())
			}

			if diags.ErrorsCount() != len(tc.wantErrors) || strings.Join(got, ",") != strings.Join(tc.wantErrors, ",") {
				t.Errorf("expected errors %v, got %v", tc.wantErrors, got)
			}
		})
	}
}

// providerConfigValue builds the raw provider configuration ValidateConfig is
// handed, with every attribute not in values null.
func providerConfigValue(t *testing.T, values map[string]tftypes.Value) tfsdk.Config {
	t.Helper()

	var schemaResp provider.SchemaResponse

	(&UptimeKumaProvider{}).Schema(t.Context(), provider.SchemaRequest{}, &schemaResp)

	objectType, ok := schemaResp.Schema.Type().TerraformType(t.Context()).(tftypes.Object)
	if !ok {
		t.Fatal("expected the provider schema to be an object")
	}

	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attrType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attrType, nil)
	}

	maps.Copy(attributes, values)

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    tftypes.NewValue(objectType, attributes),
	}
}

// TestValidateConfig pins that ValidateConfig judges the credential shape
// with the environment defaults applied, and leaves a shape alone that still
// depends on an unknown value.
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		values    map[string]tftypes.Value
		wantError string
	}{
		{
			name:   "username in config, password in environment",
			env:    map[string]string{"UPTIMEKUMA_PASSWORD": "secret"},
			values: map[string]tftypes.Value{"username": tftypes.NewValue(tftypes.String, "admin")},
		},
		{
			name:      "username without password",
			values:    map[string]tftypes.Value{"username": tftypes.NewValue(tftypes.String, "admin")},
			wantError: "password required",
		},
		{
			name: "unknown password",
			values: map[string]tftypes.Value{
				"username": tftypes.NewValue(tftypes.String, "admin"),
				"password": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			},
		},
		{
			name:      "totp secret from environment with a session token",
			env:       map[string]string{"UPTIMEKUMA_TOTP_SECRET": "JBSWY3DPEHPK3PXP"},
			values:    map[string]tftypes.Value{"session_token": tftypes.NewValue(tftypes.String, "token")},
			wantError: "username and password required",
		},
		{
			name:      "auth scheme from environment",
			env:       map[string]string{"UPTIMEKUMA_AUTH_SCHEME": "cookie"},
			wantError: "invalid auth_scheme",
		},
		{
			// The schema validator reports a configured value; ValidateConfig
			// must not report it a second time.
			name:   "auth scheme from config is left to the schema validator",
			values: map[string]tftypes.Value{"auth_scheme": tftypes.NewValue(tftypes.String, "cookie")},
		},
		{
			name:   "auth scheme from environment, valid",
			env:    map[string]string{"UPTIMEKUMA_AUTH_SCHEME": "socket"},
			values: map[string]tftypes.Value{},
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

func TestParseAuthScheme(t *testing.T) {
	t.Parallel()

	for value, want := range map[string]kuma.AuthScheme{
		"":               kuma.AuthSchemeAuto,
		authSchemeAuto:   kuma.AuthSchemeAuto,
		authSchemeSocket: kuma.AuthSchemeSocket,
		authSchemeHTTP:   kuma.AuthSchemeHTTP,
	} {
		if got := parseAuthScheme(value); got != want {
			t.Errorf("parseAuthScheme(%q): expected %s, got %s", value, want, got)
		}

		// The attribute values are the client's own names for the schemes,
		// which the pool mismatch message prints.
		if value != "" && want.String() != value {
			t.Errorf("expected kuma.AuthScheme %d to be named %q, got %q", want, value, want.String())
		}
	}
}

// TestBuildClientConfig pins that every credential of the provider model
// reaches the client configuration.
func TestBuildClientConfig(t *testing.T) {
	t.Parallel()

	model := nullProviderModel()
	model.Endpoint = types.StringValue("http://localhost:3001")
	model.Username = types.StringValue("admin")
	model.Password = types.StringValue("secret")
	model.SessionToken = types.StringValue("token")
	model.TOTPSecret = types.StringValue("JBSWY3DPEHPK3PXP")
	model.AuthScheme = types.StringValue(authSchemeHTTP)

	config := buildClientConfig(&model, clientOptions{maxRetries: 2})

	if config.Endpoint != "http://localhost:3001" || config.Username != "admin" || config.Password != "secret" {
		t.Errorf("expected endpoint, username and password to be passed on, got %q, %q, %q",
			config.Endpoint, config.Username, config.Password)
	}

	if config.SessionToken != "token" {
		t.Errorf("expected session token %q, got %q", "token", config.SessionToken)
	}

	if config.TOTPSecret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected totp secret %q, got %q", "JBSWY3DPEHPK3PXP", config.TOTPSecret)
	}

	if config.AuthScheme != kuma.AuthSchemeHTTP {
		t.Errorf("expected auth scheme %s, got %s", kuma.AuthSchemeHTTP, config.AuthScheme)
	}

	if !config.EnableConnectionPool || config.MaxRetries != 2 {
		t.Errorf("expected pooling and max retries 2, got %t and %d", config.EnableConnectionPool, config.MaxRetries)
	}
}

// TestWarnRejectedSessionToken pins the one report that a stored token has
// died. The connection succeeds without it, so a dropped call or an inverted
// condition would leave the configuration applying until the day the password
// beside it is removed.
func TestWarnRejectedSessionToken(t *testing.T) {
	t.Parallel()

	var accepted diag.Diagnostics

	warnRejectedSessionToken(nil, &accepted)

	if count := len(accepted); count != 0 {
		t.Errorf("expected no diagnostic for an accepted token, got %d", count)
	}

	rejection := errors.New("session token rejected")

	var rejected diag.Diagnostics

	warnRejectedSessionToken(rejection, &rejected)

	if rejected.HasError() {
		t.Errorf("expected a rejected token to warn rather than fail, got %v", rejected.Errors())
	}

	warnings := rejected.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected exactly one warning, got %d", len(warnings))
	}

	if summary := warnings[0].Summary(); summary != "session token rejected, fell back to password" {
		t.Errorf("expected summary %q, got %q", "session token rejected, fell back to password", summary)
	}

	if detail := warnings[0].Detail(); !strings.Contains(detail, rejection.Error()) {
		t.Errorf("expected the warning to carry the reason %q, got %q", rejection, detail)
	}
}

// TestAccProviderTOTPSecretWithoutCredentials asserts that Terraform rejects a
// configuration the provider could not honour, before it connects: the
// one-time code a `totp_secret` produces is only ever asked for by a password
// login. It needs the Terraform CLI but no Uptime Kuma, and costs no login.
func TestAccProviderTOTPSecretWithoutCredentials(t *testing.T) {
	// The environment defaults are applied before the credentials are
	// judged, so a username and password left in the environment would turn
	// this into a valid shape.
	clearCredentialEnvironment(t)

	// Not resource.ParallelTest: t.Setenv marks this test as one that cannot
	// run in parallel, and t.Parallel would panic.
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
				ExpectError: regexp.MustCompile(`totp_secret is the secret the one-time code`),
			},
		},
	})
}

// TestAccProviderMalformedTOTPSecret asserts that a secret no code can be
// derived from is named at plan time, by the attribute, rather than reaching
// the server and being reported as a connection failure.
func TestAccProviderMalformedTOTPSecret(t *testing.T) {
	clearCredentialEnvironment(t)

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

// TestApplyEnvironmentDefaults_EmptyCredentialsFallBackToEnv pins the rule
// for every credential: an empty attribute is unset, so the environment
// applies to it, while an unknown one is left for Configure to reject.
func TestApplyEnvironmentDefaults_EmptyCredentialsFallBackToEnv(t *testing.T) {
	env := map[string]string{
		"UPTIMEKUMA_USERNAME":      "admin-from-env",
		"UPTIMEKUMA_PASSWORD":      "password-from-env",
		"UPTIMEKUMA_SESSION_TOKEN": "token-from-env",
		"UPTIMEKUMA_TOTP_SECRET":   "JBSWY3DPEHPK3PXP",
		"UPTIMEKUMA_AUTH_SCHEME":   "http",
	}
	for name, value := range env {
		t.Setenv(name, value)
	}

	attributes := func(m *UptimeKumaProviderModel) map[string]*types.String {
		return map[string]*types.String{
			"UPTIMEKUMA_USERNAME":      &m.Username,
			"UPTIMEKUMA_PASSWORD":      &m.Password,
			"UPTIMEKUMA_SESSION_TOKEN": &m.SessionToken,
			"UPTIMEKUMA_TOTP_SECRET":   &m.TOTPSecret,
			"UPTIMEKUMA_AUTH_SCHEME":   &m.AuthScheme,
		}
	}

	empty := nullProviderModel()
	for _, value := range attributes(&empty) {
		*value = types.StringValue("")
	}

	applyEnvironmentDefaults(&empty, &diag.Diagnostics{})

	for name, value := range attributes(&empty) {
		if got := value.ValueString(); got != env[name] {
			t.Errorf("expected an empty attribute to take %s %q, got %q", name, env[name], got)
		}
	}

	unknown := nullProviderModel()
	for _, value := range attributes(&unknown) {
		*value = types.StringUnknown()
	}

	applyEnvironmentDefaults(&unknown, &diag.Diagnostics{})

	for name, value := range attributes(&unknown) {
		if !value.IsUnknown() {
			t.Errorf("expected an unknown attribute not to be replaced by %s, got %q", name, value.ValueString())
		}
	}
}

func TestTOTPSecretValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{name: "null", value: types.StringNull()},
		{name: "unknown", value: types.StringUnknown()},
		{name: "empty", value: types.StringValue("")},
		{name: "valid", value: types.StringValue("jbsw y3dp-ehpk 3pxp")},
		{name: "invalid", value: types.StringValue("JBSWY3DPEHPK3PX0"), wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{Path: path.Root("totp_secret"), ConfigValue: tc.value}

			var resp validator.StringResponse

			totpSecretValidator{}.ValidateString(t.Context(), req, &resp)

			if got := resp.Diagnostics.HasError(); got != tc.wantError {
				t.Fatalf("expected error %t, got %v", tc.wantError, resp.Diagnostics.Errors())
			}

			if !tc.wantError {
				return
			}

			withPath, ok := resp.Diagnostics.Errors()[0].(diag.DiagnosticWithPath)
			if !ok || !withPath.Path().Equal(path.Root("totp_secret")) {
				t.Errorf("expected the error on totp_secret, got %v", resp.Diagnostics.Errors()[0])
			}

			if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, "base32 shared secret") {
				t.Errorf("expected the detail to name the expected format, got %q", detail)
			}
		})
	}
}

// TestConfigure_FailsBeforeConnecting pins the configurations Configure
// rejects before it connects. None of them reaches a server: the endpoint is
// a port nothing listens on, so a test that did connect would fail with a
// connection error instead of the one expected.
func TestConfigure_FailsBeforeConnecting(t *testing.T) {
	str := func(value string) tftypes.Value { return tftypes.NewValue(tftypes.String, value) }

	endpoint := str("http://127.0.0.1:1")

	tests := []struct {
		name        string
		env         map[string]string
		values      map[string]tftypes.Value
		wantSummary string
	}{
		{
			name:        "username without password",
			values:      map[string]tftypes.Value{"endpoint": endpoint, "username": str("admin")},
			wantSummary: "password required",
		},
		{
			name: "session token and totp secret without username and password",
			values: map[string]tftypes.Value{
				"endpoint": endpoint, "session_token": str("token"), "totp_secret": str("JBSWY3DPEHPK3PXP"),
			},
			wantSummary: "username and password required",
		},
		{
			// The schema validator never sees a secret from the environment;
			// client.New rejects it ahead of the pool the provider always uses.
			name: "malformed totp secret from the environment",
			env:  map[string]string{"UPTIMEKUMA_TOTP_SECRET": "JBSWY3DPEHPK3PX0"},
			values: map[string]tftypes.Value{
				"endpoint": endpoint, "username": str("admin"), "password": str("secret"),
			},
			wantSummary: "Invalid TOTP secret",
		},
		{
			name: "unknown password",
			values: map[string]tftypes.Value{
				"endpoint": endpoint, "username": str("admin"),
				"password": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			},
			wantSummary: "unknown password",
		},
		{
			name:        "auth scheme from the environment",
			env:         map[string]string{"UPTIMEKUMA_AUTH_SCHEME": "cookie"},
			values:      map[string]tftypes.Value{"endpoint": endpoint, "session_token": str("token")},
			wantSummary: "invalid auth_scheme",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearCredentialEnvironment(t)
			t.Setenv("UPTIMEKUMA_ENDPOINT", "")

			for name, value := range tc.env {
				t.Setenv(name, value)
			}

			req := provider.ConfigureRequest{Config: providerConfigValue(t, tc.values)}

			var resp provider.ConfigureResponse

			(&UptimeKumaProvider{}).Configure(t.Context(), req, &resp)

			if count := resp.Diagnostics.ErrorsCount(); count != 1 {
				t.Fatalf("expected exactly one error %q, got %v", tc.wantSummary, resp.Diagnostics.Errors())
			}

			if got := resp.Diagnostics.Errors()[0].Summary(); got != tc.wantSummary {
				t.Errorf("expected error %q, got %q: %s", tc.wantSummary, got, resp.Diagnostics.Errors()[0].Detail())
			}

			if resp.ResourceData != nil {
				t.Error("expected no provider data after a failed Configure")
			}
		})
	}
}
