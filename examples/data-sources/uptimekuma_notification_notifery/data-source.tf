# Look up an existing Notifery notification by name
data "uptimekuma_notification_notifery" "alerts" {
  name = "Notifery Alerts"
}

# Look up by ID
data "uptimekuma_notification_notifery" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_notifery.alerts.id]
}
