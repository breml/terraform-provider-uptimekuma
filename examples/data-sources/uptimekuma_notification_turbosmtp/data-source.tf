# Look up an existing TurboSMTP notification by name
data "uptimekuma_notification_turbosmtp" "alerts" {
  name = "TurboSMTP Alerts"
}

# Look up by ID
data "uptimekuma_notification_turbosmtp" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_turbosmtp.alerts.id]
}
