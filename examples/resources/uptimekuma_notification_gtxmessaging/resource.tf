# Basic GTX Messaging SMS notification
resource "uptimekuma_notification_gtxmessaging" "example" {
  name      = "GTX Messaging Alerts"
  api_key   = "your-gtx-messaging-api-key"
  from      = "UptimeKuma"
  to        = "+1234567890"
  is_active = true
}
