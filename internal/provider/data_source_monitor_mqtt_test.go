package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceMonitorMQTT(t *testing.T) {
	name := acctest.RandomWithPrefix("mqtt-datasource-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMonitorMQTTConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_mqtt.by_id", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.by_id", "name", name),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.by_id", "topic", "test/datasource"),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_monitor_mqtt.by_id", "domain_expiry_notification", "true",
					),
					resource.TestCheckResourceAttrSet("data.uptimekuma_monitor_mqtt.by_name", "id"),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.by_name", "name", name),
					resource.TestCheckResourceAttr("data.uptimekuma_monitor_mqtt.by_name", "topic", "test/datasource"),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_monitor_mqtt.by_name", "domain_expiry_notification", "true",
					),
				),
			},
		},
	})
}

func testAccDataSourceMonitorMQTTConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mqtt" "test" {
  name                       = %[1]q
  hostname                   = "localhost"
  port                       = 1883
  mqtt_topic                 = "test/datasource"
  mqtt_check_type            = "keyword"
  domain_expiry_notification = true
}

data "uptimekuma_monitor_mqtt" "by_id" {
  id = uptimekuma_monitor_mqtt.test.id
}

data "uptimekuma_monitor_mqtt" "by_name" {
  name = uptimekuma_monitor_mqtt.test.name
}
`, name)
}
