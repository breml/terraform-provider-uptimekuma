resource "uptimekuma_notification_46elks" "example" {
  name        = "46elks SMS Notification"
  is_active   = true
  username    = "your_46elks_username"
  auth_token  = "your_46elks_auth_token"
  from_number = "+1234567890"
  to_number   = "+0987654321"
}
