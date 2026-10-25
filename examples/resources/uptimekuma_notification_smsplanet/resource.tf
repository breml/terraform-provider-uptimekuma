resource "uptimekuma_notification_smsplanet" "example" {
  name          = "SMS Planet Notifications"
  api_token     = "your-smsplanet-api-token"
  phone_numbers = "+48123456789"
  sender_name   = "UptimeKuma"
  is_active     = true
  is_default    = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_token_wo is never stored in
# plan or state. Change api_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_smsplanet" "write_only" {
  name = "SMS Planet with a write-only secret"

  phone_numbers = "+48123456789"

  api_token_wo         = "your-smsplanet-api-token"
  api_token_wo_version = "1"
}
