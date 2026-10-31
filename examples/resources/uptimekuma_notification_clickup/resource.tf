resource "uptimekuma_notification_clickup" "example" {
  name         = "ClickUp Notifications"
  token        = "pk_12345678_ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"
  workspace_id = "9012345678"
  channel_id   = "abc12-345"
  is_default   = false
}

# Leave the monitor address out of the message body
resource "uptimekuma_notification_clickup" "no_url" {
  name         = "ClickUp Notifications without URL"
  token        = "pk_12345678_ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"
  workspace_id = "9012345678"
  channel_id   = "abc12-345"
  disable_url  = true
}

# Requires Terraform or OpenTofu 1.11 or later: token_wo is never stored in
# plan or state. Change token_wo_version whenever the secret changes.
resource "uptimekuma_notification_clickup" "write_only" {
  name = "ClickUp with a write-only secret"

  workspace_id = "9012345678"
  channel_id   = "4-90123456-8"

  token_wo         = "pk_your-personal-token"
  token_wo_version = "1"
}
