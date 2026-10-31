resource "uptimekuma_notification_serverchan" "example" {
  name       = "ServerChan Notification"
  is_default = false

  # ServerChan send key for authentication
  # Generate this from your ServerChan account
  send_key = "SCT12345abcdef"
}

# Requires Terraform or OpenTofu 1.11 or later: send_key_wo is never stored in
# plan or state. Change send_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_serverchan" "write_only" {
  name = "ServerChan with a write-only secret"

  send_key_wo         = "SCT12345abcdef"
  send_key_wo_version = "1"
}
