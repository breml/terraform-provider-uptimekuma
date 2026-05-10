# Kafka Producer Monitor Resource Example

# Basic Kafka Producer monitor with minimum required configuration
resource "uptimekuma_monitor_kafka_producer" "basic" {
  name    = "Kafka Producer Monitor"
  brokers = ["kafka.example.com:9092"]
  topic   = "uptime-monitor"
  message = "ping"
}

# Kafka Producer monitor connecting to a multi-broker cluster
resource "uptimekuma_monitor_kafka_producer" "cluster" {
  name = "Kafka Cluster Monitor"
  brokers = [
    "kafka-1.example.com:9092",
    "kafka-2.example.com:9092",
    "kafka-3.example.com:9092",
  ]
  topic                     = "uptime-monitor"
  message                   = "ping"
  ssl                       = true
  allow_auto_topic_creation = false
  interval                  = 120
  retry_interval            = 60
  max_retries               = 3
}

# Kafka Producer monitor with SASL authentication
resource "uptimekuma_monitor_kafka_producer" "sasl" {
  name    = "Kafka Producer SASL"
  brokers = ["kafka.example.com:9093"]
  topic   = "uptime-monitor"
  message = "ping"
  ssl     = true
  sasl_options = jsonencode({
    mechanism = "scram-sha-512"
    username  = "monitor-user"
    password  = "monitor-password"
  })

  notification_ids = [1, 2]

  tags = [
    {
      tag_id = 1
      value  = "production"
    }
  ]
}
