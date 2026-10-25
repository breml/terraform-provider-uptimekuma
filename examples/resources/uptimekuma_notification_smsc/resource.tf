resource "uptimekuma_notification_smsc" "example" {
  name        = "SMSC SMS Notification"
  is_active   = true
  login       = "your_smsc_login"
  password    = "your_smsc_password"
  to_number   = "77123456789"
  sender_name = "Uptime"
  translit    = "0"
}

# Requires Terraform or OpenTofu 1.11 or later: password_wo is never stored in
# plan or state. Change password_wo_version whenever the secret changes.
resource "uptimekuma_notification_smsc" "write_only" {
  name = "SMSC with a write-only secret"

  login     = "your_smsc_login"
  to_number = "77123456789"

  password_wo         = "your_smsc_password"
  password_wo_version = "1"
}
