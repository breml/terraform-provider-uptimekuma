resource "uptimekuma_notification_bark" "example" {
  name        = "Bark Notification"
  is_active   = true
  endpoint    = "https://api.bark.com"
  group       = "monitoring"
  sound       = "default"
  api_version = "v1"
}
