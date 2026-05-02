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

func TestAccMonitorRabbitMQResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRabbitMQMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestRabbitMQMonitorUpdated")
	nodes := `["http://rabbitmq.example.com:15672/"]`
	nodesUpdated := `["http://rabbitmq1.example.com:15672/","http://rabbitmq2.example.com:15672/"]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorRabbitMQResourceConfig(name, nodes, "guest", "guest"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("nodes"),
						knownvalue.StringExact(nodes),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("guest"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact("guest"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(48),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorRabbitMQResourceConfig(nameUpdated, nodesUpdated, "admin", "secret"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("nodes"),
						knownvalue.StringExact(nodesUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("admin"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact("secret"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorRabbitMQResourceConfig(name string, nodes string, username string, password string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_rabbitmq" "test" {
  name     = %[1]q
  nodes    = %[2]q
  username = %[3]q
  password = %[4]q
  interval = 60
  active   = true
}
`, name, nodes, username, password)
}

func TestAccMonitorRabbitMQResourceWithOptionalFields(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRabbitMQMonitorWithOptional")
	description := "Test RabbitMQ monitor with optional fields"
	nodes := `["http://rabbitmq.example.com:15672/"]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRabbitMQResourceConfigWithOptionalFields(name, description, nodes),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("nodes"),
						knownvalue.StringExact(nodes),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(30),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorRabbitMQResourceConfigWithOptionalFields(
	name string,
	description string,
	nodes string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_rabbitmq" "test" {
  name            = %[1]q
  description     = %[2]q
  nodes           = %[3]q
  timeout         = 30
  interval        = 60
  retry_interval  = 60
  resend_interval = 0
  max_retries     = 3
  upside_down     = false
  active          = true
}
`, name, description, nodes)
}

func TestAccMonitorRabbitMQResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestRabbitMQGroup")
	monitorName := acctest.RandomWithPrefix("TestRabbitMQMonitorWithParent")
	nodes := `["http://rabbitmq.example.com:15672/"]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRabbitMQResourceConfigWithParent(groupName, monitorName, nodes),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("nodes"),
						knownvalue.StringExact(nodes),
					),
				},
			},
		},
	})
}

func testAccMonitorRabbitMQResourceConfigWithParent(
	groupName string,
	monitorName string,
	nodes string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_rabbitmq" "test" {
  name   = %[2]q
  nodes  = %[3]q
  parent = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName, nodes)
}

func TestAccMonitorRabbitMQResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRabbitMQMonitorImport")
	nodes := `["http://rabbitmq.example.com:15672/"]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRabbitMQResourceConfig(name, nodes, "guest", "guest"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_rabbitmq.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
