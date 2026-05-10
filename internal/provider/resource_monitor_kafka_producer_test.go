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

func TestAccMonitorKafkaProducerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestKafkaProducerMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestKafkaProducerMonitorUpdated")
	description := "Test Kafka Producer monitor description"
	descriptionUpdated := "Updated test Kafka Producer monitor description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorKafkaProducerResourceConfig(
					name,
					description,
					"kafka-1.example.com:9092",
					"test-topic",
					"hello world",
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka-1.example.com:9092"),
						}),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("test-topic"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("message"),
						knownvalue.StringExact("hello world"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("ssl"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("allow_auto_topic_creation"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				// Refresh-only step: ensure no perpetual diff is produced
				// after refreshing the existing configuration from the API.
				RefreshState:       true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccMonitorKafkaProducerResourceConfig(
					nameUpdated,
					descriptionUpdated,
					"kafka-2.example.com:9093",
					"updated-topic",
					"updated message",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(descriptionUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka-2.example.com:9093"),
						}),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("updated-topic"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("message"),
						knownvalue.StringExact("updated message"),
					),
				},
			},
			{
				ResourceName:                         "uptimekuma_monitor_kafka_producer.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "id",
				ImportStateVerifyIgnore:              []string{"sasl_options", "message"},
			},
		},
	})
}

func testAccMonitorKafkaProducerResourceConfig(
	name string,
	description string,
	broker string,
	topic string,
	message string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_kafka_producer" "test" {
  name        = %[1]q
  description = %[2]q
  brokers     = [%[3]q]
  topic       = %[4]q
  message     = %[5]q
  active      = true
}
`, name, description, broker, topic, message)
}

func TestAccMonitorKafkaProducerResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestKafkaProducerMonitorMinimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorKafkaProducerResourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka.example.com:9092"),
						}),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("monitor-topic"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("message"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("ssl"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("allow_auto_topic_creation"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorKafkaProducerResourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_kafka_producer" "test" {
  name    = %[1]q
  brokers = ["kafka.example.com:9092"]
  topic   = "monitor-topic"
  message = "ping"
}
`, name)
}

func TestAccMonitorKafkaProducerResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestKafkaProducerMonitorFull")
	description := "Full test Kafka Producer monitor"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorKafkaProducerResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("brokers"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("kafka-1.example.com:9092"),
							knownvalue.StringExact("kafka-2.example.com:9092"),
						}),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("topic"),
						knownvalue.StringExact("monitor-topic"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("message"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("ssl"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("allow_auto_topic_creation"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorKafkaProducerResourceConfigWithAllOptions(name string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_kafka_producer" "test" {
  name                      = %[1]q
  description               = %[2]q
  brokers                   = ["kafka-1.example.com:9092", "kafka-2.example.com:9092"]
  topic                     = "monitor-topic"
  message                   = "ping"
  ssl                       = true
  allow_auto_topic_creation = true
  sasl_options              = jsonencode({ mechanism = "plain", username = "user", password = "pass" })
  interval                  = 120
  retry_interval            = 60
  max_retries               = 5
  active                    = true
  upside_down               = false
}
`, name, description)
}

func TestAccMonitorKafkaProducerResourceWithParent(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestKafkaProducerGroup")
	monitorName := acctest.RandomWithPrefix("TestKafkaProducerMonitorWithParent")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorKafkaProducerResourceConfigWithParent(groupName, monitorName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(groupName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(monitorName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_kafka_producer.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccMonitorKafkaProducerResourceConfigWithParent(groupName string, monitorName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_monitor_kafka_producer" "test" {
  name    = %[2]q
  brokers = ["kafka.example.com:9092"]
  topic   = "monitor-topic"
  message = "ping"
  parent  = uptimekuma_monitor_group.test.id
}
`, groupName, monitorName)
}
