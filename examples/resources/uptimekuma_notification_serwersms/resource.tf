resource "uptimekuma_notification_serwersms" "example" {
  name         = "SerwerSMS Notification"
  is_active    = true
  username     = "your_serwersms_username"
  password     = "your_serwersms_password"
  phone_number = "+48501234567"
  sender_name  = "YourCompany"
}
