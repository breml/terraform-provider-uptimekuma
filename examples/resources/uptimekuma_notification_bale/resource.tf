resource "uptimekuma_notification_bale" "example" {
  name       = "Bale Notifications"
  bot_token  = "123456:ABCDEF"
  chat_id    = "111222333"
  is_active  = true
  is_default = false
}
