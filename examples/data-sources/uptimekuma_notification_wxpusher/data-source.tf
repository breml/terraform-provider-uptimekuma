# Look up an existing WxPusher notification by name
data "uptimekuma_notification_wxpusher" "alerts" {
  name = "WxPusher Alerts"
}

# Look up by ID
data "uptimekuma_notification_wxpusher" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_wxpusher.alerts.id]
}
