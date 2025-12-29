resource "uptimekuma_notification_wecom" "example" {
  name      = "WeCom Notifications"
  bot_key   = "your-bot-key"
  is_active = true
  is_default = false
}

resource "uptimekuma_notification_wecom" "alerts" {
  name      = "WeCom Alerts"
  bot_key   = "your-bot-key-for-alerts"
  is_active = true
  is_default = false
}
