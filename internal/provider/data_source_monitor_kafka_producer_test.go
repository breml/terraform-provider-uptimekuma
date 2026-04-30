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

func TestAccMonitorKafkaProducerDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestKafkaProducerMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorKafkaProducerDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka.example.com:9092"),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("monitor-topic"),
					),
				},
			},
			{
				Config: testAccMonitorKafkaProducerDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka.example.com:9092"),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("monitor-topic"),
					),
				},
			},
		},
	})
}

func testAccMonitorKafkaProducerDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_kafka_producer" "test" {
  name    = %[1]q
  brokers = ["kafka.example.com:9092"]
  topic   = "monitor-topic"
  message = "ping"
}

data "uptimekuma_monitor_kafka_producer" "test" {
  name = uptimekuma_monitor_kafka_producer.test.name
}
`, name)
}

func testAccMonitorKafkaProducerDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_kafka_producer" "test" {
  name    = %[1]q
  brokers = ["kafka.example.com:9092"]
  topic   = "monitor-topic"
  message = "ping"
}

data "uptimekuma_monitor_kafka_producer" "test" {
  id = uptimekuma_monitor_kafka_producer.test.id
}
`, name)
}
