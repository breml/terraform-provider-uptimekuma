# Get RabbitMQ monitor by name
data "uptimekuma_monitor_rabbitmq" "example" {
  name = "RabbitMQ Health Check"
}

# Get RabbitMQ monitor by ID
data "uptimekuma_monitor_rabbitmq" "example_by_id" {
  id = 42
}

# Use data source to reference an existing monitor
data "uptimekuma_monitor_rabbitmq" "existing" {
  name = "Production RabbitMQ"
}

output "monitor_id" {
  description = "ID of the RabbitMQ monitor"
  value       = data.uptimekuma_monitor_rabbitmq.example.id
}

output "monitor_name" {
  description = "Name of the RabbitMQ monitor"
  value       = data.uptimekuma_monitor_rabbitmq.example.name
}
