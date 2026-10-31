resource "uptimekuma_notification_onechat" "example" {
  name         = "OneChat Notifications"
  access_token = "your-access-token"
  receiver_id  = "user123"
  bot_id       = "bot-001"
  is_default   = false
}

# Requires Terraform or OpenTofu 1.11 or later: access_token_wo is never stored in
# plan or state. Change the matching *_wo_version whenever a secret changes.
resource "uptimekuma_notification_onechat" "write_only" {
  name        = "OneChat with write-only secret"
  receiver_id = "user123"
  bot_id      = "bot-001"

  access_token_wo         = "your-access-token"
  access_token_wo_version = "1"
}
