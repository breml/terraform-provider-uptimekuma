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

func TestAccMonitorTailscalePingDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestTailscalePingMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorTailscalePingDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_tailscale_ping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorTailscalePingDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_tailscale_ping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorTailscalePingDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_tailscale_ping" "test" {
  name     = %[1]q
  hostname = "100.64.0.1"
}

data "uptimekuma_monitor_tailscale_ping" "test" {
  name = uptimekuma_monitor_tailscale_ping.test.name
}
`, name)
}

func testAccMonitorTailscalePingDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_tailscale_ping" "test" {
  name     = %[1]q
  hostname = "100.64.0.1"
}

data "uptimekuma_monitor_tailscale_ping" "test" {
  id = uptimekuma_monitor_tailscale_ping.test.id
}
`, name)
}
