resource "uptimekuma_notification_46elks" "example" {
  name        = "46elks SMS Notification"
  is_active   = true
  username    = "your_46elks_username"
  auth_token  = "your_46elks_auth_token"
  from_number = "+1234567890"
  to_number   = "+0987654321"
}

# Requires Terraform or OpenTofu 1.11 or later: auth_token_wo is never stored in
# plan or state. Change auth_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_46elks" "write_only" {
  name = "46elks with a write-only secret"

  username    = "elks-user"
  from_number = "UptimeKuma"
  to_number   = "+46700000000"

  auth_token_wo         = "your-46elks-auth-token"
  auth_token_wo_version = "1"
}
