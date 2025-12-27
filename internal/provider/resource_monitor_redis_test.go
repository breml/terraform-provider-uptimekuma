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

func TestAccMonitorRedisResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRedisMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestRedisMonitorUpdated")
	connectionString := "redis://user:password@localhost:6379"
	connectionStringUpdated := "redis://user:password@localhost:6380"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorRedisResourceConfig(name, connectionString, false),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("ignore_tls"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorRedisResourceConfig(nameUpdated, connectionStringUpdated, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionStringUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("ignore_tls"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorRedisResourceConfig(name string, connectionString string, ignoreTLS bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_redis" "test" {
  name                        = %[1]q
  database_connection_string  = %[2]q
  ignore_tls                  = %[3]t
  interval                    = 60
  active                      = true
}
`, name, connectionString, ignoreTLS)
}

func TestAccMonitorRedisResourceWithOptionalFields(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRedisMonitorWithOptional")
	description := "Test Redis monitor with optional fields"
	connectionString := "redis://user:password@localhost:6379"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRedisResourceConfigWithOptionalFields(name, description, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("ignore_tls"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorRedisResourceConfigWithOptionalFields(name string, description string, connectionString string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_redis" "test" {
  name                       = %[1]q
  description                = %[2]q
  database_connection_string = %[3]q
  ignore_tls                 = false
  interval                   = 60
  retry_interval             = 60
  resend_interval            = 0
  max_retries                = 3
  upside_down                = false
  active                     = true
}
`, name, description, connectionString)
}

func TestAccMonitorRedisResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestRedisGroup")
	monitorName := acctest.RandomWithPrefix("TestRedisMonitorWithParent")
	connectionString := "redis://user:password@localhost:6379"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRedisResourceConfigWithParent(groupName, monitorName, connectionString),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_redis.test",
						tfjsonpath.New("database_connection_string"),
						knownvalue.StringExact(connectionString),
					),
				},
			},
		},
	})
}

func testAccMonitorRedisResourceConfigWithParent(groupName string, monitorName string, connectionString string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_redis" "test" {
  name                       = %[2]q
  database_connection_string = %[3]q
  parent                     = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName, connectionString)
}
