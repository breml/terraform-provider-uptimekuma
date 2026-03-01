package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
