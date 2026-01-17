# Look up an existing Nextcloud Talk notification by name
data "uptimekuma_notification_nextcloudtalk" "alerts" {
  name = "Nextcloud Alerts"
}

# Look up by ID
data "uptimekuma_notification_nextcloudtalk" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_nextcloudtalk.alerts.id]
}
