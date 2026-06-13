# Example: Read System Service monitor by name
data "uptimekuma_monitor_system_service" "example" {
  name = "Nginx Service"
}

# Example: Read System Service monitor by ID
data "uptimekuma_monitor_system_service" "by_id" {
  id = 1
}
