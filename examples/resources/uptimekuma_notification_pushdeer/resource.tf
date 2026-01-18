# PushDeer notification resource with default server
resource "uptimekuma_notification_pushdeer" "example" {
  name      = "PushDeer Notification"
  key       = "pushkey123"
  is_active = true
}

# PushDeer notification resource with custom server
resource "uptimekuma_notification_pushdeer" "custom_server" {
  name      = "PushDeer Custom Server"
  key       = "pushkey456"
  server    = "https://custom.pushdeer.server.com"
  is_active = true
}
