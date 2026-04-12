resource "uptimekuma_notification_smspartner" "example" {
  name         = "SMS Partner Alerts"
  api_key      = "YOUR_SMSPARTNER_API_KEY"
  phone_number = "+33612345678"
  sender_name  = "UptimeKuma"
  is_active    = true
  is_default   = false
}

resource "uptimekuma_notification_smspartner" "production" {
  name         = "SMS Partner Production Alerts"
  api_key      = "your_api_key_here"
  phone_number = "+33698765432"
  is_active    = true
  is_default   = true
}
