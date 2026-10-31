resource "uptimekuma_notification_wecom" "example" {
  name       = "WeCom Notifications"
  bot_key    = "your-bot-key"
  is_default = false
}

resource "uptimekuma_notification_wecom" "alerts" {
  name       = "WeCom Alerts"
  bot_key    = "your-bot-key-for-alerts"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: bot_key_wo is never stored in
# plan or state. Change bot_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_wecom" "write_only" {
  name = "WeCom with a write-only secret"

  bot_key_wo         = "your-bot-key"
  bot_key_wo_version = "1"
}
