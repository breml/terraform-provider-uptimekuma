resource "uptimekuma_notification_notifery" "example" {
  name       = "Notifery Notifications"
  api_key    = "your-notifery-api-key"
  title      = "Uptime Kuma Alert"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_notifery" "with_group" {
  name       = "Notifery with Group"
  api_key    = "your-notifery-api-key"
  title      = "Uptime Kuma Alert"
  group      = "infrastructure"
  is_active  = true
  is_default = false
}
