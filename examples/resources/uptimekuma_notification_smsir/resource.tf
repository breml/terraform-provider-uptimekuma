resource "uptimekuma_notification_smsir" "example" {
  name       = "SMS.ir Notifications"
  api_key    = "your-smsir-api-key"
  number     = "09123456789"
  template   = "123456"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_smsir" "multiple_recipients" {
  name       = "SMS.ir Alerts"
  api_key    = "your-smsir-api-key"
  number     = "09123456789,09987654321"
  template   = "123456"
  is_active  = true
  is_default = false
}
