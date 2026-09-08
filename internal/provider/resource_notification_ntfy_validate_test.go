package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestValidateNtfyAuthentication covers the credential validation for every authentication
// method, including the case that hits
// https://github.com/breml/terraform-provider-uptimekuma/issues/389: an omitted
// authentication_method is null during validation, because schema defaults are only applied
// when the plan is created.
func TestValidateNtfyAuthentication(t *testing.T) {
	tests := []struct {
		name                 string
		authenticationMethod types.String
		accessToken          types.String
		username             types.String
		password             types.String
		expectedErrors       []string
	}{
		{
			name:                 "method omitted, no credentials",
			authenticationMethod: types.StringNull(),
		},
		{
			name:                 "method omitted, access token set",
			authenticationMethod: types.StringNull(),
			accessToken:          types.StringValue("tk_token"),
			expectedErrors:       []string{"access_token"},
		},
		{
			name:                 "method omitted, username and password set",
			authenticationMethod: types.StringNull(),
			username:             types.StringValue("user"),
			password:             types.StringValue("secret"),
			expectedErrors:       []string{"username", "password"},
		},
		{
			name:                 "method unknown, credentials set",
			authenticationMethod: types.StringUnknown(),
			accessToken:          types.StringValue("tk_token"),
		},
		{
			name:                 "none, no credentials",
			authenticationMethod: types.StringValue("none"),
		},
		{
			name:                 "none, all credentials set",
			authenticationMethod: types.StringValue("none"),
			accessToken:          types.StringValue("tk_token"),
			username:             types.StringValue("user"),
			password:             types.StringValue("secret"),
			expectedErrors:       []string{"access_token", "username", "password"},
		},
		{
			name:                 "none, empty username set",
			authenticationMethod: types.StringValue("none"),
			username:             types.StringValue(""),
			expectedErrors:       []string{"username"},
		},
		{
			name:                 "accessToken, access token set",
			authenticationMethod: types.StringValue("accessToken"),
			accessToken:          types.StringValue("tk_token"),
		},
		{
			name:                 "accessToken, access token unknown",
			authenticationMethod: types.StringValue("accessToken"),
			accessToken:          types.StringUnknown(),
		},
		{
			name:                 "accessToken, access token missing",
			authenticationMethod: types.StringValue("accessToken"),
			expectedErrors:       []string{"access_token"},
		},
		{
			name:                 "accessToken, access token empty",
			authenticationMethod: types.StringValue("accessToken"),
			accessToken:          types.StringValue(""),
			expectedErrors:       []string{"access_token"},
		},
		{
			name:                 "accessToken, username and password also set",
			authenticationMethod: types.StringValue("accessToken"),
			accessToken:          types.StringValue("tk_token"),
			username:             types.StringValue("user"),
			password:             types.StringValue("secret"),
			expectedErrors:       []string{"username", "password"},
		},
		{
			name:                 "accessToken, username only known after apply",
			authenticationMethod: types.StringValue("accessToken"),
			accessToken:          types.StringValue("tk_token"),
			username:             types.StringUnknown(),
			expectedErrors:       []string{"username"},
		},
		{
			name:                 "usernamePassword, username and password set",
			authenticationMethod: types.StringValue("usernamePassword"),
			username:             types.StringValue("user"),
			password:             types.StringValue("secret"),
		},
		{
			name:                 "usernamePassword, password missing",
			authenticationMethod: types.StringValue("usernamePassword"),
			username:             types.StringValue("user"),
			expectedErrors:       []string{"password"},
		},
		{
			name:                 "usernamePassword, both missing",
			authenticationMethod: types.StringValue("usernamePassword"),
			expectedErrors:       []string{"username", "password"},
		},
		{
			name:                 "usernamePassword, access token also set",
			authenticationMethod: types.StringValue("usernamePassword"),
			username:             types.StringValue("user"),
			password:             types.StringValue("secret"),
			accessToken:          types.StringValue("tk_token"),
			expectedErrors:       []string{"access_token"},
		},
		{
			name:                 "unexpected method is left to the schema validator",
			authenticationMethod: types.StringValue("bogus"),
			accessToken:          types.StringValue("tk_token"),
			username:             types.StringValue("user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NotificationNtfyResourceModel{
				AccessToken:          tt.accessToken,
				AuthenticationMethod: tt.authenticationMethod,
				Username:             tt.username,
				Password:             tt.password,
			}

			resp := &resource.ValidateConfigResponse{}

			validateNtfyAuthentication(&config, resp)

			gotPaths := make([]string, 0, len(resp.Diagnostics))

			for _, d := range resp.Diagnostics {
				withPath, ok := d.(diag.DiagnosticWithPath)
				if !ok {
					t.Fatalf("diagnostic %q has no attribute path", d.Summary())
				}

				gotPaths = append(gotPaths, withPath.Path().String())
			}

			if len(gotPaths) != len(tt.expectedErrors) {
				t.Fatalf("got errors for %v, want errors for %v", gotPaths, tt.expectedErrors)
			}

			for i, want := range tt.expectedErrors {
				if gotPaths[i] != want {
					t.Errorf("error %d is for %q, want %q", i, gotPaths[i], want)
				}
			}
		})
	}
}
