resource "uptimekuma_notification_onesender" "private" {
  name          = "OneSender Private"
  url           = "https://onesender.example.com/api/v1/send"
  token         = "your-api-token"
  receiver      = "+6281234567890"
  type_receiver = "private"
  is_active     = true
  is_default    = false
}

resource "uptimekuma_notification_onesender" "group" {
  name          = "OneSender Group"
  url           = "https://onesender.example.com/api/v1/send"
  token         = "your-api-token"
  receiver      = "group-id-123"
  type_receiver = "group"
  is_active     = true
  is_default    = false
}

# Requires Terraform or OpenTofu 1.11 or later: token_wo is never stored in plan
# or state. Change the matching *_wo_version whenever a secret changes.
resource "uptimekuma_notification_onesender" "write_only" {
  name          = "OneSender with write-only secret"
  url           = "https://onesender.example.com/api/v1/send"
  receiver      = "+6281234567890"
  type_receiver = "private"

  token_wo         = "your-api-token"
  token_wo_version = "1"
}
