package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMonitorMQTTResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMQTTMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestMQTTMonitorUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMonitorMQTTResourceConfig(name, "localhost", 1883, "test/topic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("uptimekuma_monitor_mqtt.test", "id"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "name", name),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "hostname", "localhost"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "port", "1883"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "mqtt_topic", "test/topic"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "mqtt_check_type", "keyword"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "uptimekuma_monitor_mqtt.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"mqtt_password", // Sensitive field not returned from API
				},
			},
			// Update and Read testing
			{
				Config: testAccMonitorMQTTResourceConfig(nameUpdated, "localhost", 1883, "updated/topic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "name", nameUpdated),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "mqtt_topic", "updated/topic"),
				),
			},
		},
	})
}

func TestAccMonitorMQTTResourceWithAuth(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMQTTMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with authentication
			{
				Config: testAccMonitorMQTTResourceConfigWithAuth(
					name,
					"localhost",
					1883,
					"auth/topic",
					"testuser",
					"testpass",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("uptimekuma_monitor_mqtt.test", "id"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "name", name),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "mqtt_username", "testuser"),
				),
			},
		},
	})
}

func testAccMonitorMQTTResourceConfig(
	name string,
	hostname string,
	port int64,
	topic string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mqtt" "test" {
  name            = %[1]q
  hostname        = %[2]q
  port            = %[3]d
  mqtt_topic      = %[4]q
  mqtt_check_type = "keyword"
}
`, name, hostname, port, topic)
}

func testAccMonitorMQTTResourceConfigWithAuth(
	name string,
	hostname string,
	port int64,
	topic string,
	username string,
	password string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mqtt" "test" {
  name             = %[1]q
  hostname         = %[2]q
  port             = %[3]d
  mqtt_topic       = %[4]q
  mqtt_username    = %[5]q
  mqtt_password    = %[6]q
  mqtt_check_type  = "keyword"
}
`, name, hostname, port, topic, username, password)
}
