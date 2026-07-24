# Look up an existing Teltonika notification by name
data "uptimekuma_notification_teltonika" "alerts" {
  name = "Teltonika SMS"
}

# Look up by ID
data "uptimekuma_notification_teltonika" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_teltonika.alerts.id]
}
