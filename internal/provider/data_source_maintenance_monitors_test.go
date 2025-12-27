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

func TestAccMaintenanceMonitorsDataSource(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	monitorName1 := acctest.RandomWithPrefix("TestMonitor1")
	monitorName2 := acctest.RandomWithPrefix("TestMonitor2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceMonitorsDataSourceConfig(maintenanceTitle, monitorName1, monitorName2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance_monitors.test",
						tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
		},
	})
}

func testAccMaintenanceMonitorsDataSourceConfig(
	maintenanceTitle string,
	monitorName1 string,
	monitorName2 string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_monitor_http" "test1" {
  name = %[2]q
  url  = "https://example.com"
}

resource "uptimekuma_monitor_http" "test2" {
  name = %[3]q
  url  = "https://example.org"
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = [
    uptimekuma_monitor_http.test1.id,
    uptimekuma_monitor_http.test2.id,
  ]
}

data "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  depends_on     = [uptimekuma_maintenance_monitors.test]
}
`, maintenanceTitle, monitorName1, monitorName2)
}
