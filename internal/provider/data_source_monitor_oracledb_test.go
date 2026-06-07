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

func TestAccMonitorOracleDBDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestOracleDBMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorOracleDBDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorOracleDBDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorOracleDBDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_oracledb" "test" {
  name                       = %[1]q
  database_connection_string = "localhost:1521/ORCL"
}

data "uptimekuma_monitor_oracledb" "test" {
  name = uptimekuma_monitor_oracledb.test.name
}
`, name)
}

func testAccMonitorOracleDBDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_oracledb" "test" {
  name                       = %[1]q
  database_connection_string = "localhost:1521/ORCL"
}

data "uptimekuma_monitor_oracledb" "test" {
  id = uptimekuma_monitor_oracledb.test.id
}
`, name)
}
