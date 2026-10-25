resource "uptimekuma_notification_smspartner" "example" {
  name         = "SMS Partner Alerts"
  api_key      = "YOUR_SMSPARTNER_API_KEY"
  phone_number = "+33612345678"
  sender_name  = "UptimeKuma"
  is_active    = true
  is_default   = false
}

resource "uptimekuma_notification_smspartner" "production" {
  name         = "SMS Partner Production Alerts"
  api_key      = "your_api_key_here"
  phone_number = "+33698765432"
  is_active    = true
  is_default   = true
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_smspartner" "write_only" {
  name = "SMS Partner with a write-only secret"

  phone_number = "+33612345678"

  api_key_wo         = "your_smspartner_api_key"
  api_key_wo_version = "1"
}
