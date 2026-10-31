resource "uptimekuma_notification_kook" "example" {
  name       = "Kook Notifications"
  bot_token  = "1/MzAxMjk5NzA1OTMzMDAwMA=="
  guild_id   = "382941547624206336"
  is_default = false
}

resource "uptimekuma_notification_kook" "alerts" {
  name       = "Kook Alerts"
  bot_token  = "1/MzAxMjk5NzA1OTMzMDAwMA=="
  guild_id   = "382941547624206337"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: bot_token_wo is never stored in
# plan or state. Change bot_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_kook" "write_only" {
  name     = "Kook with write-only bot token"
  guild_id = "382941547624206338"

  bot_token_wo         = "1/MzAxMjk5NzA1OTMzMDAwMA=="
  bot_token_wo_version = "1"
}
