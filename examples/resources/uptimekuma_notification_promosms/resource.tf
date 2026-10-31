resource "uptimekuma_notification_promosms" "example" {
  name           = "PromoSMS Notification"
  login          = "your_promosms_login"
  password       = "your_promosms_password"
  phone_number   = "+48501234567"
  sender_name    = "YourCompany"
  sms_type       = "1"
  allow_long_sms = false
}

# Requires Terraform or OpenTofu 1.11 or later: password_wo is never stored in
# plan or state. Change password_wo_version whenever the secret changes.
resource "uptimekuma_notification_promosms" "write_only" {
  name = "PromoSMS with a write-only secret"

  login        = "your_promosms_login"
  phone_number = "+48501234567"

  password_wo         = "your-promosms-password"
  password_wo_version = "1"
}
