resource "uptimekuma_notification_onechat" "example" {
  name         = "OneChat Notifications"
  access_token = "your-access-token"
  receiver_id  = "user123"
  bot_id       = "bot-001"
  is_active    = true
  is_default   = false
}
