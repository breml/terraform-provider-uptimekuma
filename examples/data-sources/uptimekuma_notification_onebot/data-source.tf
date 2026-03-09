# Look up an existing OneBot notification by name
data "uptimekuma_notification_onebot" "alerts" {
  name = "OneBot Alerts"
}

# Look up by ID
data "uptimekuma_notification_onebot" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_onebot.alerts.id]
}
