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

func TestAccMonitorMongoDBDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMongoDBMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorMongoDBDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorMongoDBDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorMongoDBDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mongodb" "test" {
  name                       = %[1]q
  database_connection_string = "mongodb://user:password@localhost:27017/db"
}

data "uptimekuma_monitor_mongodb" "test" {
  name = uptimekuma_monitor_mongodb.test.name
}
`, name)
}

func testAccMonitorMongoDBDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mongodb" "test" {
  name                       = %[1]q
  database_connection_string = "mongodb://user:password@localhost:27017/db"
}

data "uptimekuma_monitor_mongodb" "test" {
  id = uptimekuma_monitor_mongodb.test.id
}
`, name)
}
