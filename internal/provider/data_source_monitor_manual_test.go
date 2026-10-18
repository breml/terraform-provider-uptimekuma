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

// TestAccMonitorManualDataSource verifies both lookup forms of the data source.
func TestAccMonitorManualDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestManualMonitor")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorManualDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_manual.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_manual.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_manual.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorManualDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_manual" "test" {
  name          = %[1]q
  manual_status = "up"
}

data "uptimekuma_monitor_manual" "by_name" {
  name = uptimekuma_monitor_manual.test.name
}

data "uptimekuma_monitor_manual" "by_id" {
  id = uptimekuma_monitor_manual.test.id
}
`, name)
}

// TestAccMonitorManualDataSourceWrongType verifies that looking up a monitor
// of another type by ID is reported. The manual monitor has no type-specific
// fields it could fail to decode, so any monitor would otherwise match.
func TestAccMonitorManualDataSourceWrongType(t *testing.T) {
	name := acctest.RandomWithPrefix("TestManualMonitorDSWrongType")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorManualDataSourceConfigWrongType(name),
				ExpectError: regexp.MustCompile(`Monitor type mismatch`),
			},
		},
	})
}

func testAccMonitorManualDataSourceConfigWrongType(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_pm2" "test" {
  name         = %[1]q
  process_name = "api-server"
}

data "uptimekuma_monitor_manual" "by_id" {
  id = uptimekuma_monitor_pm2.test.id
}
`, name)
}
