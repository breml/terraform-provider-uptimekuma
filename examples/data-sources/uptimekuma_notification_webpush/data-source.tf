# Look up an existing Web Push notification by name
data "uptimekuma_notification_webpush" "alerts" {
  name = "Web Push Notifications"
}

# Look up by ID
data "uptimekuma_notification_webpush" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_webpush.alerts.id]
}
