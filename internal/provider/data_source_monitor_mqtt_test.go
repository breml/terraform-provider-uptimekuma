package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceMonitorMQTTByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMonitorMQTTByIDConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_mqtt.test", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.test", "name", "mqtt-datasource-test"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.test", "topic", "test/datasource"),
				),
			},
		},
	})
}

func TestAccDataSourceMonitorMQTTByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMonitorMQTTByNameConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_mqtt.test", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.test", "name", "mqtt-datasource-test"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.test", "topic", "test/datasource"),
				),
			},
		},
	})
}

func testAccDataSourceMonitorMQTTByIDConfig() string {
	return providerConfig() + `
resource "uptimekuma_monitor_mqtt" "test" {
  name            = "mqtt-datasource-test"
  hostname        = "localhost"
  port            = 1883
  mqtt_topic      = "test/datasource"
  mqtt_check_type = "keyword"
}

data "uptimekuma_monitor_mqtt" "test" {
  id = uptimekuma_monitor_mqtt.test.id
}
`
}

func testAccDataSourceMonitorMQTTByNameConfig() string {
	return providerConfig() + `
resource "uptimekuma_monitor_mqtt" "test" {
  name            = "mqtt-datasource-test"
  hostname        = "localhost"
  port            = 1883
  mqtt_topic      = "test/datasource"
  mqtt_check_type = "keyword"
}

data "uptimekuma_monitor_mqtt" "test" {
  name = uptimekuma_monitor_mqtt.test.name
}
`
}
