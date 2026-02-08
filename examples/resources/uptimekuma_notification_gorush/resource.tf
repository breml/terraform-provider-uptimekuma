# Basic Gorush notification with required fields
resource "uptimekuma_notification_gorush" "basic" {
  name         = "Basic Gorush Notification"
  server_url   = "https://gorush.example.com"
  device_token = "your-device-token-here"
  platform     = "ios"
  is_active    = true
  is_default   = false
}

# Gorush notification with all optional fields
resource "uptimekuma_notification_gorush" "full" {
  name         = "Full Gorush Notification"
  server_url   = "https://gorush.example.com"
  device_token = "your-device-token-here"
  platform     = "android"
  title        = "Uptime Kuma Alert"
  priority     = "high"
  retry        = 3
  topic        = "uptime-alerts"
  is_active    = true
  is_default   = false
}

# Gorush notification for web push
resource "uptimekuma_notification_gorush" "web" {
  name         = "Web Push Notification"
  server_url   = "https://gorush.example.com"
  device_token = "web-device-subscription-key"
  platform     = "web"
  title        = "Status Update"
  is_active    = true
}
