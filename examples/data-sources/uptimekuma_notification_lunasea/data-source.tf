# Look up an existing Lunasea notification by name
data "uptimekuma_notification_lunasea" "alerts" {
  name = "Lunasea Notifications"
}

# Look up by ID
data "uptimekuma_notification_lunasea" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_lunasea.alerts.id]
}
