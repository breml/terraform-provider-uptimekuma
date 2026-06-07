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

func TestAccMonitorOracleDBResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestOracleDBMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestOracleDBMonitorUpdated")
	connectionString := "localhost:1521/ORCL"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorOracleDBResourceConfig(name, connectionString, "SELECT 1 FROM DUAL"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT 1 FROM DUAL"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorOracleDBResourceConfig(nameUpdated, connectionString, "SELECT SYSDATE FROM DUAL"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT SYSDATE FROM DUAL"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorOracleDBResourceConfig(name string, connectionString string, query string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_oracledb" "test" {
  name                        = %[1]q
  database_connection_string  = %[2]q
  database_query              = %[3]q
  interval                    = 60
  active                      = true
}
`, name, connectionString, query)
}

func TestAccMonitorOracleDBResourceWithOptionalFields(t *testing.T) {
	name := acctest.RandomWithPrefix("TestOracleDBMonitorWithOptional")
	description := "Test OracleDB monitor with optional fields"
	connectionString := "localhost:1521/ORCL"
	username := "monitoring_user"
	password := "s3cr3t"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorOracleDBResourceConfigWithOptionalFields(
					name,
					description,
					connectionString,
					username,
					password,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT 1 FROM DUAL"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("basic_auth_user"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorOracleDBResourceConfigWithOptionalFields(
	name string,
	description string,
	connectionString string,
	username string,
	password string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_oracledb" "test" {
  name                       = %[1]q
  description                = %[2]q
  database_connection_string = %[3]q
  basic_auth_user            = %[4]q
  basic_auth_pass            = %[5]q
  interval                   = 60
  retry_interval             = 60
  resend_interval            = 0
  max_retries                = 3
  upside_down                = false
  active                     = true
}
`, name, description, connectionString, username, password)
}

func TestAccMonitorOracleDBResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestOracleDBGroup")
	monitorName := acctest.RandomWithPrefix("TestOracleDBMonitorWithParent")
	connectionString := "localhost:1521/ORCL"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorOracleDBResourceConfigWithParent(groupName, monitorName, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
				},
			},
		},
	})
}

func testAccMonitorOracleDBResourceConfigWithParent(
	groupName string,
	monitorName string,
	connectionString string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_oracledb" "test" {
  name                       = %[2]q
  database_connection_string = %[3]q
  parent                     = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName, connectionString)
}

func TestAccMonitorOracleDBResourceWithConditions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestOracleDBConditions")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorOracleDBResourceConfigWithConditions(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("conditions"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("variable"),
						knownvalue.StringExact("result"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("operator"),
						knownvalue.StringExact("contains"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_oracledb.test",
						tfjsonpath.New("conditions").AtSliceIndex(1).AtMapKey("and_or"),
						knownvalue.StringExact("or"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_monitor_oracledb.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_connection_string", "basic_auth_pass"},
			},
		},
	})
}

func testAccMonitorOracleDBResourceConfigWithConditions(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_oracledb" "test" {
  name                       = %[1]q
  database_connection_string = "localhost:1521/ORCL"
  database_query             = "SELECT COUNT(*) FROM dual"

  conditions = [
    {
      variable = "result"
      operator = "contains"
      value    = "1"
    },
    {
      variable = "result"
      operator = "contains"
      value    = "0"
      and_or   = "or"
    },
  ]
}
`, name)
}
