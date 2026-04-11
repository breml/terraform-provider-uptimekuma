resource "uptimekuma_notification_smsmanager" "example" {
  name         = "SMS Manager Notification"
  is_active    = true
  api_key      = "your_smsmanager_api_key"
  numbers      = "+1234567890"
  message_type = "sms"
}
