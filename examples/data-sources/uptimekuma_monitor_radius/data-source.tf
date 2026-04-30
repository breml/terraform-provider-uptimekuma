# Radius Monitor Data Source Example

# Reference an existing Radius monitor by ID
data "uptimekuma_monitor_radius" "by_id" {
  id = 42
}

# Reference an existing Radius monitor by name
data "uptimekuma_monitor_radius" "by_name" {
  name = "Radius Authentication Monitor"
}

# Use data source output in another resource
resource "uptimekuma_notification_webhook" "radius_webhook" {
  name = "Radius Monitor Notifications"
  url  = "https://example.com/notify"
}

# Create a notification association using data source
resource "uptimekuma_monitor_radius" "monitored" {
  name             = "Production Radius Server"
  hostname         = "prod-radius.internal"
  radius_username  = "monitor-user"
  radius_password  = "monitor-password"
  radius_secret    = "shared-secret"
  notification_ids = [uptimekuma_notification_webhook.radius_webhook.id]
}
