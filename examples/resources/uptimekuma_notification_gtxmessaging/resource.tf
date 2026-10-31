# Basic GTX Messaging SMS notification
resource "uptimekuma_notification_gtxmessaging" "example" {
  name    = "GTX Messaging Alerts"
  api_key = "your-gtx-messaging-api-key"
  from    = "UptimeKuma"
  to      = "+1234567890"
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_gtxmessaging" "write_only" {
  name = "GTX Messaging with a write-only secret"

  from = "UptimeKuma"
  to   = "+41790000000"

  api_key_wo         = "your-gtx-api-key"
  api_key_wo_version = "1"
}
