# Look up an existing Free Mobile notification by name
data "uptimekuma_notification_freemobile" "alerts" {
  name = "Free Mobile SMS Alerts"
}

# Look up by ID
data "uptimekuma_notification_freemobile" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_freemobile.alerts.id]
}
