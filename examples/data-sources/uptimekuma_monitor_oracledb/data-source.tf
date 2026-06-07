# Get OracleDB monitor by name
data "uptimekuma_monitor_oracledb" "example" {
  name = "Oracle Database Health Check"
}

# Get OracleDB monitor by ID
data "uptimekuma_monitor_oracledb" "example_by_id" {
  id = 42
}

# Use data source to reference an existing monitor
data "uptimekuma_monitor_oracledb" "existing" {
  name = "Production Oracle"
}

output "monitor_id" {
  description = "ID of the OracleDB monitor"
  value       = data.uptimekuma_monitor_oracledb.example.id
}

output "monitor_name" {
  description = "Name of the OracleDB monitor"
  value       = data.uptimekuma_monitor_oracledb.example.name
}
