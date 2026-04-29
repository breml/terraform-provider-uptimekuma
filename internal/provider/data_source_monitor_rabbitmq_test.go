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

func TestAccMonitorRabbitMQDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRabbitMQMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRabbitMQDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorRabbitMQDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_rabbitmq.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorRabbitMQDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_rabbitmq" "test" {
  name  = %[1]q
  nodes = "[\"http://rabbitmq.example.com:15672/\"]"
}

data "uptimekuma_monitor_rabbitmq" "test" {
  name = uptimekuma_monitor_rabbitmq.test.name
}
`, name)
}

func testAccMonitorRabbitMQDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_rabbitmq" "test" {
  name  = %[1]q
  nodes = "[\"http://rabbitmq.example.com:15672/\"]"
}

data "uptimekuma_monitor_rabbitmq" "test" {
  id = uptimekuma_monitor_rabbitmq.test.id
}
`, name)
}
