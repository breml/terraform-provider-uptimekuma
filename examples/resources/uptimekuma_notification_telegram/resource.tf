resource "uptimekuma_notification_telegram" "example" {
  name    = "Telegram Notifications"
  bot_token = "YOUR_BOT_TOKEN"
  chat_id = "YOUR_CHAT_ID"
  is_active = true
  is_default = false
}

resource "uptimekuma_notification_telegram" "with_options" {
  name                = "Telegram with Options"
  bot_token           = "YOUR_BOT_TOKEN"
  chat_id             = "YOUR_CHAT_ID"
  send_silently       = true
  protect_content     = true
  use_template        = false
  template_parse_mode = "HTML"
  is_active           = true
  is_default          = false
}

resource "uptimekuma_notification_telegram" "with_custom_server" {
  name      = "Telegram with Custom Server"
  bot_token = "YOUR_BOT_TOKEN"
  chat_id   = "YOUR_CHAT_ID"
  server_url = "https://custom-telegram-api.example.com"
  is_active = true
  is_default = false
}
