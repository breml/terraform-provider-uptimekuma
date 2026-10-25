resource "uptimekuma_notification_serwersms" "example" {
  name         = "SerwerSMS Notification"
  is_active    = true
  username     = "your_serwersms_username"
  password     = "your_serwersms_password"
  phone_number = "+48501234567"
  sender_name  = "YourCompany"
}

# Requires Terraform or OpenTofu 1.11 or later: password_wo is never stored in
# plan or state. Change password_wo_version whenever the secret changes.
resource "uptimekuma_notification_serwersms" "write_only" {
  name = "SerwerSMS with a write-only secret"

  username     = "your_serwersms_username"
  phone_number = "+48501234567"
  sender_name  = "YourCompany"

  password_wo         = "your_serwersms_password"
  password_wo_version = "1"
}
