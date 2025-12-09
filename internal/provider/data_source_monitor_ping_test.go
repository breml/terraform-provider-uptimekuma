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

func TestAccMonitorPingDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPingMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPingDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_ping.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
			{
				Config: testAccMonitorPingDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_ping.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccMonitorPingDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = "8.8.8.8"
}

data "uptimekuma_monitor_ping" "test" {
  name = uptimekuma_monitor_ping.test.name
}
`, name)
}

func testAccMonitorPingDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = "8.8.8.8"
}

data "uptimekuma_monitor_ping" "test" {
  id = uptimekuma_monitor_ping.test.id
}
`, name)
}
