# Look up an existing DingDing notification by name
data "uptimekuma_notification_dingding" "alerts" {
  name = "DingDing Notifications"
}

# Look up by ID
data "uptimekuma_notification_dingding" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_dingding.alerts.id]
}
