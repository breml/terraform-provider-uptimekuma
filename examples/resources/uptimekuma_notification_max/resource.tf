resource "uptimekuma_notification_max" "example" {
  name      = "MAX Notifications"
  bot_token = "your-bot-token"
  chat_id   = "-123456789"
  is_active = true

  # Optional: Use a custom message template
  use_template    = true
  template        = "Monitor: {{ name }}\nStatus: {{ status }}"
  template_format = "markdown"
}
