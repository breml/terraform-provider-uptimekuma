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

func TestAccMonitorPostgresDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPostgresMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPostgresDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_postgres.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
			{
				Config: testAccMonitorPostgresDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_postgres.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccMonitorPostgresDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_postgres" "test" {
  name                       = %[1]q
  database_connection_string = "postgres://user:password@localhost:5432/db"
}

data "uptimekuma_monitor_postgres" "test" {
  name = uptimekuma_monitor_postgres.test.name
}
`, name)
}

func testAccMonitorPostgresDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_postgres" "test" {
  name                       = %[1]q
  database_connection_string = "postgres://user:password@localhost:5432/db"
}

data "uptimekuma_monitor_postgres" "test" {
  id = uptimekuma_monitor_postgres.test.id
}
`, name)
}
