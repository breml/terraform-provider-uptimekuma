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

func TestAccMonitorMQTTResourceWithJSONQuery(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMQTTMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with json-query check type
			{
				Config: testAccMonitorMQTTResourceConfigWithJSONQuery(
					name,
					"localhost",
					1883,
					"json/topic",
					"$.status",
					"active",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("uptimekuma_monitor_mqtt.test", "id"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "name", name),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "mqtt_check_type", "json-query"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "json_path", "$.status"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "expected_value", "active"),
				),
			},
		},
	})
}

func TestAccMonitorMQTTResourceWithWebSocket(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMQTTMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with WebSocket configuration
			{
				Config: testAccMonitorMQTTResourceConfigWithWebSocket(
					name,
					"localhost",
					"ws/topic",
					"/mqtt",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("uptimekuma_monitor_mqtt.test", "id"),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "name", name),
					resource.TestCheckResourceAttr("uptimekuma_monitor_mqtt.test", "mqtt_websocket_path", "/mqtt"),
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

func testAccMonitorMQTTResourceConfigWithJSONQuery(
	name string,
	hostname string,
	port int64,
	topic string,
	jsonPath string,
	expectedValue string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mqtt" "test" {
  name             = %[1]q
  hostname         = %[2]q
  port             = %[3]d
  mqtt_topic       = %[4]q
  mqtt_check_type  = "json-query"
  json_path        = %[5]q
  expected_value   = %[6]q
}
`, name, hostname, port, topic, jsonPath, expectedValue)
}

func testAccMonitorMQTTResourceConfigWithWebSocket(
	name string,
	hostname string,
	topic string,
	websocketPath string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mqtt" "test" {
  name                 = %[1]q
  hostname             = %[2]q
  mqtt_topic           = %[3]q
  mqtt_websocket_path  = %[4]q
  mqtt_check_type      = "keyword"
}
`, name, hostname, topic, websocketPath)
}

func TestAccMonitorMQTTResourceWithConditions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestMQTTConditions")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorMQTTResourceConfigWithConditions(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("variable"),
						knownvalue.StringExact("topic"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("operator"),
						knownvalue.StringExact("=="),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("value"),
						knownvalue.StringExact("sensors/temp"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(0).AtMapKey("and_or"),
						knownvalue.StringExact("and"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(1).AtMapKey("variable"),
						knownvalue.StringExact("message"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(1).AtMapKey("operator"),
						knownvalue.StringExact("contains"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(1).AtMapKey("value"),
						knownvalue.StringExact("alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_mqtt.test",
						tfjsonpath.New("conditions").AtSliceIndex(1).AtMapKey("and_or"),
						knownvalue.StringExact("or"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_mqtt.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorMQTTResourceConfigWithConditions(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_mqtt" "test" {
  name            = %[1]q
  hostname        = "localhost"
  port            = 1883
  mqtt_topic      = "sensors/temp"
  mqtt_check_type = "keyword"

  conditions = [
    {
      variable = "topic"
      operator = "=="
      value    = "sensors/temp"
    },
    {
      variable = "message"
      operator = "contains"
      value    = "alert"
      and_or   = "or"
    },
  ]
}
`, name)
}
