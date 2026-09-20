# Look up an existing Flowtriq notification by name
data "uptimekuma_notification_flowtriq" "alerts" {
  name = "Flowtriq Alerts"
}

# Look up by ID
data "uptimekuma_notification_flowtriq" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_flowtriq.alerts.id]
}
