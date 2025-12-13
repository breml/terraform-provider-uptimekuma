# Look up a monitor group by name
data "uptimekuma_monitor_group" "production" {
  name = "Production Services"
}

# Look up by ID
data "uptimekuma_monitor_group" "by_id" {
  id = 10
}

# Use the group as a parent for new monitors
resource "uptimekuma_monitor_http" "service" {
  name   = "Backend Service"
  url    = "https://backend.example.com/health"
  parent = data.uptimekuma_monitor_group.production.id
}
