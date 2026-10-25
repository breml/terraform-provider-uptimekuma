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

# Requires Terraform or OpenTofu 1.11 or later: bot_token_wo is never stored in
# plan or state. Change bot_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_max" "write_only" {
  name    = "MAX with write-only bot token"
  chat_id = "-123456789"

  bot_token_wo         = "your-bot-token"
  bot_token_wo_version = "1"
}
