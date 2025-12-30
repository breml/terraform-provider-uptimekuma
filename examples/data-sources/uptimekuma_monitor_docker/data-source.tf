# Example: Read Docker monitor by name
data "uptimekuma_monitor_docker" "example" {
  name = "Redis Container"
}

# Example: Read Docker monitor by ID
data "uptimekuma_monitor_docker" "by_id" {
  id = 1
}
