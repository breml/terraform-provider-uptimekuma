package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMonitorSQLServerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSQLServerMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestSQLServerMonitorUpdated")
	connectionString := "Server=localhost;User=sa;Password=MyPassword123;TrustServerCertificate=true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorSQLServerResourceConfig(name, connectionString, "SELECT 1"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT 1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorSQLServerResourceConfig(nameUpdated, connectionString, "SELECT @@version"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT @@version"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_monitor_sqlserver.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"database_connection_string"},
				ImportStateIdFunc:       testAccMonitorSQLServerImportStateID,
			},
		},
	})
}

func testAccMonitorSQLServerResourceConfig(name string, connectionString string, query string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sqlserver" "test" {
  name                        = %[1]q
  database_connection_string  = %[2]q
  database_query              = %[3]q
  interval                    = 60
  active                      = true
}
`, name, connectionString, query)
}

func TestAccMonitorSQLServerResourceWithOptionalFields(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSQLServerMonitorWithOptional")
	description := "Test SQL Server monitor with optional fields"
	connectionString := "Server=localhost;User=sa;Password=MyPassword123;TrustServerCertificate=true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSQLServerResourceConfigWithOptionalFields(name, description, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("database_query"),
						knownvalue.StringExact("SELECT 1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorSQLServerResourceConfigWithOptionalFields(
	name string,
	description string,
	connectionString string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sqlserver" "test" {
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

func TestAccMonitorSQLServerResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestSQLServerGroup")
	monitorName := acctest.RandomWithPrefix("TestSQLServerMonitorWithParent")
	connectionString := "Server=localhost;User=sa;Password=MyPassword123;TrustServerCertificate=true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSQLServerResourceConfigWithParent(groupName, monitorName, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sqlserver.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccMonitorSQLServerResourceConfigWithParent(
	groupName string,
	monitorName string,
	connectionString string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_sqlserver" "test" {
  name                       = %[2]q
  parent                     = uptimekuma_monitor_group.test.id
  database_connection_string = %[3]q
  active                     = true
}
`, groupName, monitorName, connectionString)
}

func testAccMonitorSQLServerImportStateID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["uptimekuma_monitor_sqlserver.test"]
	if !ok {
		return "", errors.New("Not found: uptimekuma_monitor_sqlserver.test")
	}

	return rs.Primary.Attributes["id"], nil
}
