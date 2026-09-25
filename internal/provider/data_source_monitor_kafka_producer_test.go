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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorKafkaProducerDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_name",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka.example.com:9092"),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_name",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("monitor-topic"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_name",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(2.5),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_id",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka.example.com:9092"),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_id",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("monitor-topic"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_kafka_producer.by_id",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(2.5),
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
  timeout = 2.5
}

data "uptimekuma_monitor_kafka_producer" "by_name" {
  name = uptimekuma_monitor_kafka_producer.test.name
}

data "uptimekuma_monitor_kafka_producer" "by_id" {
  id = uptimekuma_monitor_kafka_producer.test.id
}
`, name)
}
