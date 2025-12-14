# Look up an existing Feishu notification by name
data "uptimekuma_notification_feishu" "alerts" {
  name = "Feishu Notifications"
}

# Look up by ID
data "uptimekuma_notification_feishu" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_feishu.alerts.id]
}
