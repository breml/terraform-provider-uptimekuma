package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccMonitorPM2DataSource verifies both lookup forms of the data source.
func TestAccMonitorPM2DataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2Monitor")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPM2DataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_pm2.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_pm2.by_name",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact("api-server"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_pm2.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_pm2.by_id",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact("api-server"),
					),
				},
			},
		},
	})
}

func testAccMonitorPM2DataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_pm2" "test" {
  name         = %[1]q
  process_name = "api-server"
}

data "uptimekuma_monitor_pm2" "by_name" {
  name = uptimekuma_monitor_pm2.test.name
}

data "uptimekuma_monitor_pm2" "by_id" {
  id = uptimekuma_monitor_pm2.test.id
}
`, name)
}

// TestAccMonitorPM2DataSourceWrongType verifies that looking up a monitor of
// another type by ID is reported rather than decoded into empty PM2 fields. The
// system-service monitor shares the `system_service_name` wire field, so it
// would otherwise decode into a plausible-looking process name.
func TestAccMonitorPM2DataSourceWrongType(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2MonitorDSWrongType")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorPM2DataSourceConfigWrongType(name),
				ExpectError: regexp.MustCompile(`Monitor type mismatch`),
			},
		},
	})
}

func testAccMonitorPM2DataSourceConfigWrongType(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_system_service" "test" {
  name                = %[1]q
  system_service_name = "nginx.service"
}

data "uptimekuma_monitor_pm2" "by_id" {
  id = uptimekuma_monitor_system_service.test.id
}
`, name)
}
