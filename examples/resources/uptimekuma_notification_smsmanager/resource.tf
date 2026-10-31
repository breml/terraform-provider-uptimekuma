resource "uptimekuma_notification_smsmanager" "example" {
  name         = "SMS Manager Notification"
  api_key      = "your_smsmanager_api_key"
  numbers      = "+1234567890"
  message_type = "sms"
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_smsmanager" "write_only" {
  name = "SMS Manager with a write-only secret"

  numbers = "+1234567890"

  api_key_wo         = "your_smsmanager_api_key"
  api_key_wo_version = "1"
}
