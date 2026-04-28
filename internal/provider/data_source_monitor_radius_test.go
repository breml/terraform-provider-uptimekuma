package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMonitorRadiusDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRadiusMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRadiusDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.test",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("testuser"),
					),
				},
			},
			{
				Config: testAccMonitorRadiusDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.test",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("testuser"),
					),
				},
			},
		},
	})
}

func testAccMonitorRadiusDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_radius" "test" {
  name            = %[1]q
  hostname        = "radius.example.com"
  radius_username = "testuser"
  radius_password = "testpass"
  radius_secret   = "testsecret"
}

data "uptimekuma_monitor_radius" "test" {
  name = uptimekuma_monitor_radius.test.name
}
`, name)
}

func testAccMonitorRadiusDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_radius" "test" {
  name            = %[1]q
  hostname        = "radius.example.com"
  radius_username = "testuser"
  radius_password = "testpass"
  radius_secret   = "testsecret"
}

data "uptimekuma_monitor_radius" "test" {
  id = uptimekuma_monitor_radius.test.id
}
`, name)
}
