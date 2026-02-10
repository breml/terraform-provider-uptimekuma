# Look up an existing GTX Messaging notification by name
data "uptimekuma_notification_gtxmessaging" "alerts" {
  name = "GTX Messaging Alerts"
}

# Look up by ID
data "uptimekuma_notification_gtxmessaging" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_gtxmessaging.alerts.id]
}
