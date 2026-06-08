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

func TestAccMonitorSystemServiceDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSystemServiceMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSystemServiceDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_system_service.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_system_service.test",
						tfjsonpath.New("system_service_name"),
						knownvalue.StringExact("nginx.service"),
					),
				},
			},
			{
				Config: testAccMonitorSystemServiceDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_system_service.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_system_service.test",
						tfjsonpath.New("system_service_name"),
						knownvalue.StringExact("nginx.service"),
					),
				},
			},
		},
	})
}

func testAccMonitorSystemServiceDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_system_service" "test" {
  name                = %[1]q
  system_service_name = "nginx.service"
}

data "uptimekuma_monitor_system_service" "test" {
  name = uptimekuma_monitor_system_service.test.name
}
`, name)
}

func testAccMonitorSystemServiceDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_system_service" "test" {
  name                = %[1]q
  system_service_name = "nginx.service"
}

data "uptimekuma_monitor_system_service" "test" {
  id = uptimekuma_monitor_system_service.test.id
}
`, name)
}
