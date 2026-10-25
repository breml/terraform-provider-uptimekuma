resource "uptimekuma_notification_clicksendsms" "example" {
  name        = "ClickSend SMS Notification"
  is_active   = true
  login       = "your-clicksend-username"
  password    = "your-clicksend-api-key"
  to_number   = "+61412345678"
  sender_name = "UptimeKuma"
}

# Requires Terraform or OpenTofu 1.11 or later: password_wo is never stored in
# plan or state. Change password_wo_version whenever the secret changes.
resource "uptimekuma_notification_clicksendsms" "write_only" {
  name = "ClickSend SMS with a write-only secret"

  login       = "clicksend-user"
  to_number   = "+61400000000"
  sender_name = "UptimeKuma"

  password_wo         = "your-clicksend-api-key"
  password_wo_version = "1"
}
