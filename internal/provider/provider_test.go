package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){ //nolint:gochecknoglobals // Ideomatic Terraform provider code.
	"uptimekuma": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(_ *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func TestApplyEnvironmentDefaults(t *testing.T) {
	tests := []struct {
		name             string
		initial          UptimeKumaProviderModel
		envEndpoint      string
		envUsername      string
		envPassword      string
		expectedEndpoint string
		expectedUsername string
		expectedPassword string
	}{
		{
			name: "no env vars set, config null",
			initial: UptimeKumaProviderModel{
				Endpoint: types.StringNull(),
				Username: types.StringNull(),
				Password: types.StringNull(),
			},
			expectedEndpoint: "",
			expectedUsername: "",
			expectedPassword: "",
		},
		{
			name: "env vars set, config null",
			initial: UptimeKumaProviderModel{
				Endpoint: types.StringNull(),
				Username: types.StringNull(),
				Password: types.StringNull(),
			},
			envEndpoint:      "http://localhost:3001",
			envUsername:      "admin",
			envPassword:      "password",
			expectedEndpoint: "http://localhost:3001",
			expectedUsername: "admin",
			expectedPassword: "password",
		},
		{
			name: "env vars set, config has endpoint",
			initial: UptimeKumaProviderModel{
				Endpoint: types.StringValue("http://override:3001"),
				Username: types.StringNull(),
				Password: types.StringNull(),
			},
			envEndpoint:      "http://localhost:3001",
			envUsername:      "admin",
			envPassword:      "password",
			expectedEndpoint: "http://override:3001",
			expectedUsername: "admin",
			expectedPassword: "password",
		},
		{
			name: "env vars set, config has all values",
			initial: UptimeKumaProviderModel{
				Endpoint: types.StringValue("http://override:3001"),
				Username: types.StringValue("override"),
				Password: types.StringValue("override"),
			},
			envEndpoint:      "http://localhost:3001",
			envUsername:      "admin",
			envPassword:      "password",
			expectedEndpoint: "http://override:3001",
			expectedUsername: "override",
			expectedPassword: "override",
		},
		{
			name: "only env endpoint set",
			initial: UptimeKumaProviderModel{
				Endpoint: types.StringNull(),
				Username: types.StringNull(),
				Password: types.StringNull(),
			},
			envEndpoint:      "http://localhost:3001",
			expectedEndpoint: "http://localhost:3001",
			expectedUsername: "",
			expectedPassword: "",
		},
		{
			name: "partial config with env vars",
			initial: UptimeKumaProviderModel{
				Endpoint: types.StringValue("http://override:3001"),
				Username: types.StringNull(),
				Password: types.StringNull(),
			},
			envEndpoint:      "http://localhost:3001",
			envUsername:      "admin",
			envPassword:      "password",
			expectedEndpoint: "http://override:3001",
			expectedUsername: "admin",
			expectedPassword: "password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			if tc.envEndpoint != "" {
				t.Setenv("UPTIMEKUMA_ENDPOINT", tc.envEndpoint)
			}

			if tc.envUsername != "" {
				t.Setenv("UPTIMEKUMA_USERNAME", tc.envUsername)
			}

			if tc.envPassword != "" {
				t.Setenv("UPTIMEKUMA_PASSWORD", tc.envPassword)
			}

			// Apply defaults
			applyEnvironmentDefaults(&tc.initial)

			// Verify endpoint
			if tc.expectedEndpoint == "" {
				if !tc.initial.Endpoint.IsNull() && tc.initial.Endpoint.ValueString() != "" {
					t.Errorf("expected endpoint to be null or empty, got %q", tc.initial.Endpoint.ValueString())
				}
			} else {
				if tc.initial.Endpoint.ValueString() != tc.expectedEndpoint {
					t.Errorf("endpoint mismatch: got %q, want %q", tc.initial.Endpoint.ValueString(), tc.expectedEndpoint)
				}
			}

			// Verify username
			if tc.expectedUsername == "" {
				if !tc.initial.Username.IsNull() && tc.initial.Username.ValueString() != "" {
					t.Errorf("expected username to be null or empty, got %q", tc.initial.Username.ValueString())
				}
			} else {
				if tc.initial.Username.ValueString() != tc.expectedUsername {
					t.Errorf("username mismatch: got %q, want %q", tc.initial.Username.ValueString(), tc.expectedUsername)
				}
			}

			// Verify password
			if tc.expectedPassword == "" {
				if !tc.initial.Password.IsNull() && tc.initial.Password.ValueString() != "" {
					t.Errorf("expected password to be null or empty, got %q", tc.initial.Password.ValueString())
				}
			} else {
				if tc.initial.Password.ValueString() != tc.expectedPassword {
					t.Errorf("password mismatch: got %q, want %q", tc.initial.Password.ValueString(), tc.expectedPassword)
				}
			}
		})
	}
}

func TestEnvironmentVariablePrecedence(t *testing.T) {
	tests := []struct {
		name                   string
		envEndpoint            string
		envUsername            string
		envPassword            string
		configEndpoint         string
		configUsername         string
		configPassword         string
		expectedEndpoint       string
		expectedUsername       string
		expectedPassword       string
		expectConfigToOverride bool
	}{
		{
			name:             "env vars only, no config",
			envEndpoint:      "http://env:3001",
			envUsername:      "env-user",
			envPassword:      "env-pass",
			expectedEndpoint: "http://env:3001",
			expectedUsername: "env-user",
			expectedPassword: "env-pass",
		},
		{
			name:                   "config overrides endpoint from env",
			envEndpoint:            "http://env:3001",
			envUsername:            "env-user",
			envPassword:            "env-pass",
			configEndpoint:         "http://config:3001",
			expectedEndpoint:       "http://config:3001",
			expectedUsername:       "env-user",
			expectedPassword:       "env-pass",
			expectConfigToOverride: true,
		},
		{
			name:                   "config overrides all env vars",
			envEndpoint:            "http://env:3001",
			envUsername:            "env-user",
			envPassword:            "env-pass",
			configEndpoint:         "http://config:3001",
			configUsername:         "config-user",
			configPassword:         "config-pass",
			expectedEndpoint:       "http://config:3001",
			expectedUsername:       "config-user",
			expectedPassword:       "config-pass",
			expectConfigToOverride: true,
		},
		{
			name:             "only endpoint from env, no credentials",
			envEndpoint:      "http://env:3001",
			expectedEndpoint: "http://env:3001",
			expectedUsername: "",
			expectedPassword: "",
		},
		{
			name:                   "config overrides username, keeps env password",
			envEndpoint:            "http://env:3001",
			envUsername:            "env-user",
			envPassword:            "env-pass",
			configUsername:         "config-user",
			expectedEndpoint:       "http://env:3001",
			expectedUsername:       "config-user",
			expectedPassword:       "env-pass",
			expectConfigToOverride: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			if tc.envEndpoint != "" {
				t.Setenv("UPTIMEKUMA_ENDPOINT", tc.envEndpoint)
			} else {
				t.Setenv("UPTIMEKUMA_ENDPOINT", "")
			}

			if tc.envUsername != "" {
				t.Setenv("UPTIMEKUMA_USERNAME", tc.envUsername)
			} else {
				t.Setenv("UPTIMEKUMA_USERNAME", "")
			}

			if tc.envPassword != "" {
				t.Setenv("UPTIMEKUMA_PASSWORD", tc.envPassword)
			} else {
				t.Setenv("UPTIMEKUMA_PASSWORD", "")
			}

			// Create config model with values
			model := UptimeKumaProviderModel{
				Endpoint: types.StringNull(),
				Username: types.StringNull(),
				Password: types.StringNull(),
			}

			if tc.configEndpoint != "" {
				model.Endpoint = types.StringValue(tc.configEndpoint)
			}

			if tc.configUsername != "" {
				model.Username = types.StringValue(tc.configUsername)
			}

			if tc.configPassword != "" {
				model.Password = types.StringValue(tc.configPassword)
			}

			// Apply environment defaults
			applyEnvironmentDefaults(&model)

			// Verify results
			if tc.expectedEndpoint == "" {
				if !model.Endpoint.IsNull() && model.Endpoint.ValueString() != "" {
					t.Errorf("endpoint: expected empty, got %q", model.Endpoint.ValueString())
				}
			} else {
				if model.Endpoint.ValueString() != tc.expectedEndpoint {
					t.Errorf("endpoint: got %q, want %q", model.Endpoint.ValueString(), tc.expectedEndpoint)
				}
			}

			if tc.expectedUsername == "" {
				if !model.Username.IsNull() && model.Username.ValueString() != "" {
					t.Errorf("username: expected empty, got %q", model.Username.ValueString())
				}
			} else {
				if model.Username.ValueString() != tc.expectedUsername {
					t.Errorf("username: got %q, want %q", model.Username.ValueString(), tc.expectedUsername)
				}
			}

			if tc.expectedPassword == "" {
				if !model.Password.IsNull() && model.Password.ValueString() != "" {
					t.Errorf("password: expected empty, got %q", model.Password.ValueString())
				}
			} else {
				if model.Password.ValueString() != tc.expectedPassword {
					t.Errorf("password: got %q, want %q", model.Password.ValueString(), tc.expectedPassword)
				}
			}
		})
	}
}

func TestAccProviderEnvironmentVariables(t *testing.T) {
	envEndpoint := os.Getenv("UPTIMEKUMA_ENDPOINT")

	if envEndpoint == "" {
		t.Skip("UPTIMEKUMA_ENDPOINT not set - skipping environment variable provider test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderWithEnvironmentVariables(),
			},
		},
	})
}

func TestAccProviderMixedConfiguration(t *testing.T) {
	configEndpoint := os.Getenv("UPTIMEKUMA_ENDPOINT")
	if configEndpoint == "" {
		t.Skip("UPTIMEKUMA_ENDPOINT not set - skipping provider test")
	}

	originalUsername := os.Getenv("UPTIMEKUMA_USERNAME")
	originalPassword := os.Getenv("UPTIMEKUMA_PASSWORD")

	t.Setenv("UPTIMEKUMA_USERNAME", "env-user")
	t.Setenv("UPTIMEKUMA_PASSWORD", "env-pass")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "uptimekuma" {
  endpoint = %[1]q
  username = %[2]q
  password = %[3]q
}

data "uptimekuma_tag" "test" {}
`, configEndpoint, originalUsername, originalPassword),
			},
		},
	})
}

func testAccProviderWithEnvironmentVariables() string {
	return `
provider "uptimekuma" {
}

data "uptimekuma_tag" "test" {}
`
}
