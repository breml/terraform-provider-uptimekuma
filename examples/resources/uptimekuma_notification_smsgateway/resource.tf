resource "uptimekuma_notification_smsgateway" "example" {
  name       = "SMS Gateway Notifications"
  url        = "http://sms-gateway.example.com:8080"
  api_key    = "your-sms-gateway-api-key"
  recipients = "+15551234567"
  is_active  = true
  is_default = false
}

# Deliver to several recipients by separating their numbers with commas
resource "uptimekuma_notification_smsgateway" "team" {
  name       = "SMS Gateway Team Notifications"
  url        = "http://sms-gateway.example.com:8080"
  api_key    = "your-sms-gateway-api-key"
  recipients = "+15551234567, +15559876543"
  is_active  = true
  is_default = false
}
