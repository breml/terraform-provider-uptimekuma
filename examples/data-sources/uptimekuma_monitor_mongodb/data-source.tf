# Get MongoDB monitor by name
data "uptimekuma_monitor_mongodb" "example" {
  name = "MongoDB Database Health Check"
}

# Get MongoDB monitor by ID
data "uptimekuma_monitor_mongodb" "example_by_id" {
  id = 42
}

# Use data source to reference an existing monitor
data "uptimekuma_monitor_mongodb" "existing" {
  name = "Production MongoDB"
}

output "monitor_id" {
  description = "ID of the MongoDB monitor"
  value       = data.uptimekuma_monitor_mongodb.example.id
}

output "monitor_name" {
  description = "Name of the MongoDB monitor"
  value       = data.uptimekuma_monitor_mongodb.example.name
}
