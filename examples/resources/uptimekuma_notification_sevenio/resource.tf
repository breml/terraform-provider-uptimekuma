resource "uptimekuma_notification_sevenio" "example" {
  name    = "Seven.io SMS Notification"
  api_key = "your_sevenio_api_key"
  sender  = "+491234567890"
  to      = "+491111111111"
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_sevenio" "write_only" {
  name = "Seven.io with a write-only secret"

  sender = "+491234567890"
  to     = "+491111111111"

  api_key_wo         = "your_sevenio_api_key"
  api_key_wo_version = "1"
}
