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

func TestAccMonitorMongoDBResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMongoDBMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestMongoDBMonitorUpdated")
	connectionString := "mongodb://user:password@localhost:27017/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorMongoDBResourceConfig(name, connectionString, `{"ping": 1}`),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact(`{"ping": 1}`),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorMongoDBResourceConfig(nameUpdated, connectionString, `{"find": "test"}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact(`{"find": "test"}`),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorMongoDBResourceConfig(name string, connectionString string, query string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mongodb" "test" {
  name                        = %[1]q
  database_connection_string  = %[2]q
  database_query              = %[3]q
  interval                    = 60
  active                      = true
}
`, name, connectionString, query)
}

func TestAccMonitorMongoDBResourceWithOptionalFields(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMongoDBMonitorWithOptional")
	description := "Test MongoDB monitor with optional fields"
	connectionString := "mongodb://user:password@localhost:27017/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorMongoDBResourceConfigWithOptionalFields(name, description, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact(`{"ping": 1}`),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorMongoDBResourceConfigWithOptionalFields(
	name string,
	description string,
	connectionString string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mongodb" "test" {
  name                       = %[1]q
  description                = %[2]q
  database_connection_string = %[3]q
  interval                   = 60
  retry_interval             = 60
  resend_interval            = 0
  max_retries                = 3
  upside_down                = false
  active                     = true
}
`, name, description, connectionString)
}

func TestAccMonitorMongoDBResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestMongoDBGroup")
	monitorName := acctest.RandomWithPrefix("TestMongoDBMonitorWithParent")
	connectionString := "mongodb://user:password@localhost:27017/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorMongoDBResourceConfigWithParent(groupName, monitorName, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
				},
			},
		},
	})
}

func testAccMonitorMongoDBResourceConfigWithParent(
	groupName string,
	monitorName string,
	connectionString string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_mongodb" "test" {
  name                       = %[2]q
  database_connection_string = %[3]q
  parent                     = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName, connectionString)
}

func TestAccMonitorMongoDBResourceWithJSONPath(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMongoDBMonitorWithJSONPath")
	connectionString := "mongodb://user:password@localhost:27017/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorMongoDBResourceConfigWithJSONPath(name, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("json_path"),
						knownvalue.StringExact("$.status"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("expected_value"),
						knownvalue.StringExact("ok"),
					),
				},
			},
		},
	})
}

func testAccMonitorMongoDBResourceConfigWithJSONPath(name string, connectionString string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mongodb" "test" {
  name                       = %[1]q
  database_connection_string = %[2]q
  json_path                  = "$.status"
  expected_value             = "ok"
}
`, name, connectionString)
}

func TestAccMonitorMongoDBResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMongoDBMonitorImport")
	connectionString := "mongodb://user:password@localhost:27017/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorMongoDBResourceConfig(name, connectionString, `{"ping": 1}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mongodb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_mongodb.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
