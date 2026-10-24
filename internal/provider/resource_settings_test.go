package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	kuma "github.com/breml/go-uptime-kuma-client"
)

func TestAccSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccSettingsResourceConfig("Europe/Berlin", 30),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("settings"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("server_timezone"),
						knownvalue.StringExact("Europe/Berlin"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("keep_data_period_days"),
						knownvalue.Int64Exact(30),
					),
				},
			},
			{
				Config:             testAccSettingsResourceConfig("America/New_York", 7),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("server_timezone"),
						knownvalue.StringExact("America/New_York"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("keep_data_period_days"),
						knownvalue.Int64Exact(7),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Reset to defaults so subsequent tests see a clean baseline.
			{
				Config: testAccSettingsResourceConfigDefaults(),
			},
		},
	})
}

func testAccSettingsResourceConfig(timezone string, keepDays int) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_settings" "test" {
  server_timezone       = %[1]q
  keep_data_period_days = %[2]d
}
`, timezone, keepDays)
}

func TestAccSettingsResourceAllFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccSettingsResourceConfigAllFields(),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("server_timezone"),
						knownvalue.StringExact("UTC"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("keep_data_period_days"),
						knownvalue.Int64Exact(14),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("check_update"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("search_engine_index"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("entry_page"),
						knownvalue.StringExact("dashboard"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("nscd"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("tls_expiry_notify_days"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(7),
							knownvalue.Int64Exact(14),
							knownvalue.Int64Exact(21),
						}),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_settings.test",
						tfjsonpath.New("trust_proxy"),
						knownvalue.Bool(false),
					),
				},
			},
			// Reset to defaults so subsequent tests see a clean baseline.
			{
				Config: testAccSettingsResourceConfigDefaults(),
			},
		},
	})
}

func testAccSettingsResourceConfigDefaults() string {
	return providerConfig() + `
resource "uptimekuma_settings" "test" {
  server_timezone        = "UTC"
  keep_data_period_days  = 180
  check_update           = true
  search_engine_index    = true
  entry_page             = "dashboard"
  nscd                   = false
  tls_expiry_notify_days = [7, 14, 21]
  trust_proxy            = false
  primary_base_url       = ""
  steam_api_key          = ""
  chrome_executable      = ""
}
`
}

func testAccSettingsResourceConfigAllFields() string {
	return providerConfig() + `
resource "uptimekuma_settings" "test" {
  server_timezone        = "UTC"
  keep_data_period_days  = 14
  check_update           = false
  search_engine_index    = false
  entry_page             = "dashboard"
  nscd                   = true
  tls_expiry_notify_days = [7, 14, 21]
  trust_proxy            = false
}
`
}

// TestAccSettingsResourceWriteOnly runs non-parallel like every settings test, because settings
// are a singleton. It starts from a stored steam_api_key, which the write-only steps must drop
// from state, and ends by removing the write-only pair, after which Read refreshes the key from
// the server again. The last step clears the key, as destroying the resource leaves the server
// settings as they are.
func TestAccSettingsResourceWriteOnly(t *testing.T) {
	testCase := testAccWriteOnlyTestCase(t, writeOnlyTestCase{
		addr:      "uptimekuma_settings.test",
		attrs:     []string{"steam_api_key"},
		singleton: true,
		config: func(secret string, version string) string {
			return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_settings" "test" {
%s}
`, writeOnlyAttrLines(secret, version, "steam_api_key"))
		},
		serverSecrets: func(ctx context.Context, kumaClient *kuma.Client, _ int64) (map[string]string, error) {
			s, err := kumaClient.GetSettings(ctx)
			if err != nil {
				return nil, fmt.Errorf("get settings: %w", err)
			}

			return map[string]string{"steam_api_key": s.SteamAPIKey}, nil
		},
	})

	testCase.Steps = append([]resource.TestStep{{
		Config: providerConfig() + `
resource "uptimekuma_settings" "test" {
  steam_api_key = "stored"
}
`,
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("uptimekuma_settings.test", tfjsonpath.New("steam_api_key"),
				knownvalue.StringExact("stored")),
		},
	}}, testCase.Steps...)

	testCase.Steps = append(testCase.Steps, resource.TestStep{
		Config: providerConfig() + `
resource "uptimekuma_settings" "test" {}
`,
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("uptimekuma_settings.test", tfjsonpath.New("steam_api_key"),
				knownvalue.StringExact(writeOnlyTestSecret("steam_api_key", "second"))),
		},
	}, resource.TestStep{
		Config: providerConfig() + `
resource "uptimekuma_settings" "test" {
  steam_api_key = ""
}
`,
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("uptimekuma_settings.test", tfjsonpath.New("steam_api_key"),
				knownvalue.StringExact("")),
		},
	})

	resource.Test(t, testCase)
}
