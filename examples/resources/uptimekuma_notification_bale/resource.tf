resource "uptimekuma_notification_bale" "example" {
  name       = "Bale Notifications"
  bot_token  = "123456:ABCDEF"
  chat_id    = "111222333"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: bot_token_wo is never stored in
# plan or state. Change bot_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_bale" "write_only" {
  name = "Bale with a write-only secret"

  chat_id = "123456"

  bot_token_wo         = "your-bale-bot-token"
  bot_token_wo_version = "1"
}
