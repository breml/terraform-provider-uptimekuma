# Look up an existing Apprise notification by name
data "uptimekuma_notification_apprise" "alerts" {
  name = "Apprise Notifications"
}

# Look up by ID
data "uptimekuma_notification_apprise" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_apprise.alerts.id]
}
