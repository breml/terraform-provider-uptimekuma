resource "uptimekuma_notification_promosms" "example" {
  name           = "PromoSMS Notification"
  is_active      = true
  login          = "your_promosms_login"
  password       = "your_promosms_password"
  phone_number   = "+48501234567"
  sender_name    = "YourCompany"
  sms_type       = "1"
  allow_long_sms = false
}
