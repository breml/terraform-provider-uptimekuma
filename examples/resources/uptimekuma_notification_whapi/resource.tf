resource "uptimekuma_notification_whapi" "example" {
  name       = "Whapi Notifications"
  api_url    = "https://gate.whapi.cloud"
  auth_token = "your-whapi-auth-token"
  recipient  = "1234567890"
  is_active  = true
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: auth_token_wo is never stored in
# plan or state. Change auth_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_whapi" "write_only" {
  name = "Whapi with a write-only secret"

  api_url   = "https://gate.whapi.cloud"
  recipient = "1234567890"

  auth_token_wo         = "your-whapi-auth-token"
  auth_token_wo_version = "1"
}
