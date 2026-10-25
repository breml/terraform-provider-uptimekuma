package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestValidateOctopushConfig covers the field validation for both Octopush versions, including
// the omitted version, which is null during validation and means the schema default "2".
func TestValidateOctopushConfig(t *testing.T) {
	tests := []struct {
		name           string
		version        types.String
		apiKey         types.String
		apiKeyWO       types.String
		login          types.String
		phoneNumber    types.String
		dmAPIKey       types.String
		dmAPIKeyWO     types.String
		dmLogin        types.String
		dmPhoneNumber  types.String
		expectedErrors []string
	}{
		{
			name:        "version omitted, V2 fields set",
			version:     types.StringNull(),
			apiKey:      types.StringValue("key"),
			login:       types.StringValue("login"),
			phoneNumber: types.StringValue("+33600000000"),
		},
		{
			name:           "version omitted, nothing set",
			version:        types.StringNull(),
			expectedErrors: []string{"api_key", "login", "phone_number"},
		},
		{
			name:           "version omitted, V1 fields set",
			version:        types.StringNull(),
			dmAPIKey:       types.StringValue("key"),
			dmLogin:        types.StringValue("login"),
			dmPhoneNumber:  types.StringValue("+33600000000"),
			expectedErrors: []string{"api_key", "login", "phone_number", "dm_api_key", "dm_login", "dm_phone_number"},
		},
		{
			name:    "version unknown, nothing set",
			version: types.StringUnknown(),
		},
		{
			name:        "V2, write-only API key set",
			version:     types.StringValue("2"),
			apiKeyWO:    types.StringValue("key"),
			login:       types.StringValue("login"),
			phoneNumber: types.StringValue("+33600000000"),
		},
		{
			name:        "V2, API key and login unknown",
			version:     types.StringValue("2"),
			apiKey:      types.StringUnknown(),
			login:       types.StringUnknown(),
			phoneNumber: types.StringValue("+33600000000"),
		},
		{
			name:           "V2, API key empty",
			version:        types.StringValue("2"),
			apiKey:         types.StringValue(""),
			login:          types.StringValue("login"),
			phoneNumber:    types.StringValue("+33600000000"),
			expectedErrors: []string{"api_key"},
		},
		{
			name:           "V2, nothing set",
			version:        types.StringValue("2"),
			expectedErrors: []string{"api_key", "login", "phone_number"},
		},
		{
			name:           "V2, write-only V1 API key also set",
			version:        types.StringValue("2"),
			apiKey:         types.StringValue("key"),
			login:          types.StringValue("login"),
			phoneNumber:    types.StringValue("+33600000000"),
			dmAPIKeyWO:     types.StringValue("key"),
			expectedErrors: []string{"dm_api_key_wo"},
		},
		{
			name:          "V1, V1 fields set",
			version:       types.StringValue("1"),
			dmAPIKey:      types.StringValue("key"),
			dmLogin:       types.StringValue("login"),
			dmPhoneNumber: types.StringValue("+33600000000"),
		},
		{
			name:          "V1, write-only API key set",
			version:       types.StringValue("1"),
			dmAPIKeyWO:    types.StringValue("key"),
			dmLogin:       types.StringValue("login"),
			dmPhoneNumber: types.StringValue("+33600000000"),
		},
		{
			name:           "V1, nothing set",
			version:        types.StringValue("1"),
			expectedErrors: []string{"dm_api_key", "dm_login", "dm_phone_number"},
		},
		{
			name:           "V1, write-only V2 API key also set",
			version:        types.StringValue("1"),
			dmAPIKey:       types.StringValue("key"),
			dmLogin:        types.StringValue("login"),
			dmPhoneNumber:  types.StringValue("+33600000000"),
			apiKeyWO:       types.StringValue("key"),
			expectedErrors: []string{"api_key_wo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NotificationOctopushResourceModel{
				Version:       tt.version,
				APIKey:        tt.apiKey,
				APIKeyWO:      tt.apiKeyWO,
				Login:         tt.login,
				PhoneNumber:   tt.phoneNumber,
				DMAPIKey:      tt.dmAPIKey,
				DMAPIKeyWO:    tt.dmAPIKeyWO,
				DMLogin:       tt.dmLogin,
				DMPhoneNumber: tt.dmPhoneNumber,
			}

			resp := &resource.ValidateConfigResponse{}

			validateOctopushConfig(&config, resp)

			assertDiagnosticPaths(t, resp.Diagnostics, tt.expectedErrors)
		})
	}
}

// assertDiagnosticPaths checks that diags holds one error for each of want, in that order.
func assertDiagnosticPaths(t *testing.T, diags diag.Diagnostics, want []string) {
	t.Helper()

	gotPaths := make([]string, 0, len(diags))

	for _, d := range diags {
		withPath, ok := d.(diag.DiagnosticWithPath)
		if !ok {
			t.Fatalf("diagnostic %q has no attribute path", d.Summary())
		}

		gotPaths = append(gotPaths, withPath.Path().String())
	}

	if len(gotPaths) != len(want) {
		t.Fatalf("got errors for %v, want errors for %v", gotPaths, want)
	}

	for i, w := range want {
		if gotPaths[i] != w {
			t.Errorf("error %d is for %q, want %q", i, gotPaths[i], w)
		}
	}
}
