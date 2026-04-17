resource "uptimekuma_notification_wpush" "example" {
  name      = "WPush Notifications"
  api_key   = "your-wpush-api-key"
  channel   = "monitoring"
  is_active = true
}
