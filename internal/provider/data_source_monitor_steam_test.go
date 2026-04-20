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

func TestAccMonitorSteamDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSteamMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSteamDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_steam.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorSteamDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_steam.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorSteamDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_steam" "test" {
  name     = %[1]q
  hostname = "192.168.1.100"
  port     = 27015
}

data "uptimekuma_monitor_steam" "test" {
  name = uptimekuma_monitor_steam.test.name
}
`, name)
}

func testAccMonitorSteamDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_steam" "test" {
  name     = %[1]q
  hostname = "192.168.1.100"
  port     = 27015
}

data "uptimekuma_monitor_steam" "test" {
  id = uptimekuma_monitor_steam.test.id
}
`, name)
}
