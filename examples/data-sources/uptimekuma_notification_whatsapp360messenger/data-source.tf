# Look up an existing WhatsApp 360messenger notification by name
data "uptimekuma_notification_whatsapp360messenger" "alerts" {
  name = "WhatsApp 360messenger Alerts"
}

# Look up by ID
data "uptimekuma_notification_whatsapp360messenger" "by_id" {
  id = 1
}

# Use with a monitor resource
resource "uptimekuma_monitor_http" "api" {
  name             = "API Monitor"
  url              = "https://api.example.com/health"
  notification_ids = [data.uptimekuma_notification_whatsapp360messenger.alerts.id]
}
