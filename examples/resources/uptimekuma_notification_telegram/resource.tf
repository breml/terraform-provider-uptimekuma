resource "uptimekuma_notification_telegram" "example" {
  name       = "Telegram Notifications"
  bot_token  = "YOUR_BOT_TOKEN"
  chat_id    = "YOUR_CHAT_ID"
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
  is_default          = false
}

resource "uptimekuma_notification_telegram" "with_custom_server" {
  name       = "Telegram with Custom Server"
  bot_token  = "YOUR_BOT_TOKEN"
  chat_id    = "YOUR_CHAT_ID"
  server_url = "https://custom-telegram-api.example.com"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: bot_token_wo is never stored in
# plan or state. Change bot_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_telegram" "write_only" {
  name = "Telegram with a write-only secret"

  chat_id = "YOUR_CHAT_ID"

  bot_token_wo         = "YOUR_BOT_TOKEN"
  bot_token_wo_version = "1"
}
