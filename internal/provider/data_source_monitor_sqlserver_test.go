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

func TestAccMonitorSQLServerDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSQLServerMonitor")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSQLServerDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sqlserver.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sqlserver.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorSQLServerDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sqlserver" "test" {
  name                       = %[1]q
  database_connection_string = "Server=localhost;User=sa;Password=MyPassword123;TrustServerCertificate=true"
  active                     = true
}

data "uptimekuma_monitor_sqlserver" "by_name" {
  name = uptimekuma_monitor_sqlserver.test.name
}

data "uptimekuma_monitor_sqlserver" "by_id" {
  id = uptimekuma_monitor_sqlserver.test.id
}
`, name)
}
