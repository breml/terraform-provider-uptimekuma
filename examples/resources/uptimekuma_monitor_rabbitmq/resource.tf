# RabbitMQ monitor resource
resource "uptimekuma_monitor_rabbitmq" "example" {
  name     = "RabbitMQ Health Check"
  nodes    = jsonencode(["http://rabbitmq.example.com:15672/"])
  username = "guest"
  password = "guest"
  interval = 60
  active   = true
}

# RabbitMQ monitor for a multi-node cluster with custom timeout
resource "uptimekuma_monitor_rabbitmq" "example_cluster" {
  name        = "RabbitMQ Cluster"
  description = "Monitors a RabbitMQ cluster across multiple nodes"
  nodes = jsonencode([
    "http://rabbitmq1.example.com:15672/",
    "http://rabbitmq2.example.com:15672/",
    "http://rabbitmq3.example.com:15672/",
  ])
  username        = "monitoring"
  password        = "secure_password"
  timeout         = 30
  interval        = 30
  retry_interval  = 10
  resend_interval = 0
  max_retries     = 5
  upside_down     = false
  active          = true
}

# RabbitMQ monitor with tags
resource "uptimekuma_tag" "rabbitmq_environment" {
  name = "environment"
}

resource "uptimekuma_tag" "rabbitmq_service" {
  name = "service"
}

resource "uptimekuma_monitor_rabbitmq" "example_with_tags" {
  name     = "RabbitMQ with Tags"
  nodes    = jsonencode(["http://rabbitmq.example.com:15672/"])
  username = "guest"
  password = "guest"
  interval = 60
  active   = true

  tags = [
    {
      tag_id = uptimekuma_tag.rabbitmq_environment.id
      value  = "production"
    },
    {
      tag_id = uptimekuma_tag.rabbitmq_service.id
      value  = "messaging"
    },
  ]
}

# RabbitMQ monitor in a group
resource "uptimekuma_monitor_group" "messaging_monitors" {
  name = "Messaging Monitors"
}

resource "uptimekuma_monitor_rabbitmq" "example_grouped" {
  name     = "RabbitMQ in Group"
  nodes    = jsonencode(["http://rabbitmq.example.com:15672/"])
  username = "guest"
  password = "guest"
  interval = 60
  active   = true
  parent   = uptimekuma_monitor_group.messaging_monitors.id
}

# RabbitMQ monitor with notification
resource "uptimekuma_monitor_rabbitmq" "example_with_notification" {
  name     = "RabbitMQ with Alerts"
  nodes    = jsonencode(["http://rabbitmq.example.com:15672/"])
  username = "monitoring"
  password = "secure_password"
  interval = 60
  active   = true

  notification_ids = [uptimekuma_notification_slack.alerts.id]
}
