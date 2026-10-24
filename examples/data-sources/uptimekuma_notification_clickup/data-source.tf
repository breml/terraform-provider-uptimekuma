# Look up an existing ClickUp notification by name
data "uptimekuma_notification_clickup" "alerts" {
  name = "ClickUp Alerts"
}

# Look up by ID
data "uptimekuma_notification_clickup" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_clickup.alerts.id]
}
