resource "uptimekuma_notification_gotify" "example" {
  name              = "Gotify Notifications"
  is_active         = true
  is_default        = false
  apply_existing    = false
  server_url        = "https://gotify.example.com"
  application_token = "AGe0Ks4WV5fEJkX"
  priority          = 8
}
