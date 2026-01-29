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

func TestAccMonitorMySQLDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMySQLMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorMySQLDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_mysql.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorMySQLDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_mysql.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorMySQLDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mysql" "test" {
  name                       = %[1]q
  database_connection_string = "user:password@tcp(localhost:3306)/db"
}

data "uptimekuma_monitor_mysql" "test" {
  name = uptimekuma_monitor_mysql.test.name
}
`, name)
}

func testAccMonitorMySQLDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mysql" "test" {
  name                       = %[1]q
  database_connection_string = "user:password@tcp(localhost:3306)/db"
}

data "uptimekuma_monitor_mysql" "test" {
  id = uptimekuma_monitor_mysql.test.id
}
`, name)
}
