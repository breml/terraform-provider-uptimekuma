# Get MySQL monitor by name
data "uptimekuma_monitor_mysql" "example" {
  name = "MySQL Database Health Check"
}

# Get MySQL monitor by ID
data "uptimekuma_monitor_mysql" "example_by_id" {
  id = 42
}

# Use data source to reference an existing monitor
data "uptimekuma_monitor_mysql" "existing" {
  name = "Production MySQL"
}

output "monitor_id" {
  description = "ID of the MySQL monitor"
  value       = data.uptimekuma_monitor_mysql.example.id
}

output "monitor_name" {
  description = "Name of the MySQL monitor"
  value       = data.uptimekuma_monitor_mysql.example.name
}
