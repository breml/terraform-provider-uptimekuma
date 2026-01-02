resource "uptimekuma_notification_clicksendsms" "example" {
  name        = "ClickSend SMS Notification"
  is_active   = true
  login       = "your-clicksend-username"
  password    = "your-clicksend-api-key"
  to_number   = "+61412345678"
  sender_name = "UptimeKuma"
}
