resource "uptimekuma_notification_smsir" "example" {
  name       = "SMS.ir Notifications"
  api_key    = "your-smsir-api-key"
  number     = "09123456789"
  template   = "123456"
  is_default = false
}

resource "uptimekuma_notification_smsir" "multiple_recipients" {
  name       = "SMS.ir Alerts"
  api_key    = "your-smsir-api-key"
  number     = "09123456789,09987654321"
  template   = "123456"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_smsir" "write_only" {
  name = "SMS.ir with a write-only secret"

  number   = "09123456789"
  template = "123456"

  api_key_wo         = "your-smsir-api-key"
  api_key_wo_version = "1"
}
