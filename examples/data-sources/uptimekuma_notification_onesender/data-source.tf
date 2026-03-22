# Look up an existing OneSender notification by name
data "uptimekuma_notification_onesender" "alerts" {
  name = "OneSender Alerts"
}

# Look up by ID
data "uptimekuma_notification_onesender" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_onesender.alerts.id]
}
