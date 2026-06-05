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

func TestAccMonitorPostgresResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPostgresMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestPostgresMonitorUpdated")
	connectionString := "postgres://user:password@localhost:5432/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorPostgresResourceConfig(name, connectionString, "SELECT 1"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT 1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorPostgresResourceConfig(nameUpdated, connectionString, "SELECT version()"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT version()"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorPostgresResourceConfig(name string, connectionString string, query string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_postgres" "test" {
  name                        = %[1]q
  database_connection_string  = %[2]q
  database_query              = %[3]q
  interval                    = 60
  active                      = true
}
`, name, connectionString, query)
}

func TestAccMonitorPostgresResourceWithOptionalFields(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPostgresMonitorWithOptional")
	description := "Test PostgreSQL monitor with optional fields"
	connectionString := "postgres://user:password@localhost:5432/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPostgresResourceConfigWithOptionalFields(name, description, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT 1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorPostgresResourceConfigWithOptionalFields(
	name string,
	description string,
	connectionString string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_postgres" "test" {
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

func TestAccMonitorPostgresResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestPostgresGroup")
	monitorName := acctest.RandomWithPrefix("TestPostgresMonitorWithParent")
	connectionString := "postgres://user:password@localhost:5432/testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPostgresResourceConfigWithParent(groupName, monitorName, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
				},
			},
		},
	})
}

func testAccMonitorPostgresResourceConfigWithParent(
	groupName string,
	monitorName string,
	connectionString string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_postgres" "test" {
  name                       = %[2]q
  database_connection_string = %[3]q
  parent                     = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName, connectionString)
}

func TestAccMonitorPostgresResourceWithConditions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPostgresConditions")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPostgresResourceConfigWithConditions(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("conditions"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("variable"),
						knownvalue.StringExact("result"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("operator"),
						knownvalue.StringExact("contains"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_postgres.test",
						tfjsonpath.New("conditions").AtSliceIndex(1).AtMapKey("and_or"),
						knownvalue.StringExact("or"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_postgres.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorPostgresResourceConfigWithConditions(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_postgres" "test" {
  name                       = %[1]q
  database_connection_string = "postgres://user:pass@localhost:5432/db"
  database_query             = "SELECT version()"

  conditions = [
    {
      variable = "result"
      operator = "contains"
      value    = "PostgreSQL"
    },
    {
      variable = "result"
      operator = "contains"
      value    = "16"
      and_or   = "or"
    },
  ]
}
`, name)
}
