resource "uptimekuma_notification_smsplanet" "example" {
  name          = "SMS Planet Notifications"
  api_token     = "your-smsplanet-api-token"
  phone_numbers = "+48123456789"
  sender_name   = "UptimeKuma"
  is_active     = true
  is_default    = false
}
