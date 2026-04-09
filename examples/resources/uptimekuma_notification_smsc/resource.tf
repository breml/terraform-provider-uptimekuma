resource "uptimekuma_notification_smsc" "example" {
  name        = "SMSC SMS Notification"
  is_active   = true
  login       = "your_smsc_login"
  password    = "your_smsc_password"
  to_number   = "77123456789"
  sender_name = "Uptime"
  translit    = "0"
}
