# Kafka Producer Monitor Data Source Example

# Reference an existing Kafka Producer monitor by ID
data "uptimekuma_monitor_kafka_producer" "by_id" {
  id = 42
}

# Reference an existing Kafka Producer monitor by name
data "uptimekuma_monitor_kafka_producer" "by_name" {
  name = "Kafka Producer Monitor"
}

# Use data source output in another resource
resource "uptimekuma_notification_webhook" "kafka_webhook" {
  name = "Kafka Monitor Notifications"
  url  = "https://example.com/notify"
}

# Create a notification association using data source
resource "uptimekuma_monitor_kafka_producer" "monitored" {
  name             = "Production Kafka"
  brokers          = ["kafka.internal:9092"]
  topic            = "uptime-monitor"
  message          = "ping"
  notification_ids = [uptimekuma_notification_webhook.kafka_webhook.id]
}
