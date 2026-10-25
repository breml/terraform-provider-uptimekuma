resource "uptimekuma_notification_freemobile" "example" {
  name       = "Free Mobile SMS Alerts"
  user       = "YOUR_FREE_MOBILE_USER_ID"
  pass       = "YOUR_FREE_MOBILE_API_KEY"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_freemobile" "production" {
  name       = "Free Mobile Production Alerts"
  user       = "1234567890"
  pass       = "your_api_key_here"
  is_active  = true
  is_default = true
}

# Requires Terraform or OpenTofu 1.11 or later: pass_wo is never stored in
# plan or state. Change pass_wo_version whenever the secret changes.
resource "uptimekuma_notification_freemobile" "write_only" {
  name = "Free Mobile with a write-only secret"

  user = "12345678"

  pass_wo         = "your-free-mobile-api-key"
  pass_wo_version = "1"
}
