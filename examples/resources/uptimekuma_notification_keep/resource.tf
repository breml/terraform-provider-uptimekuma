resource "uptimekuma_notification_keep" "example" {
  name        = "Keep Notifications"
  is_active   = true
  webhook_url = "https://api.keephq.dev/alerts/alert"
  api_key     = "your-keep-api-key"
}
