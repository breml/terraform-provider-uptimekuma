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

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_smsgateway" "write_only" {
  name = "SMS Gateway with a write-only secret"

  url        = "http://sms-gateway.example.com:8080"
  recipients = "+15551234567"

  api_key_wo         = "your-sms-gateway-api-key"
  api_key_wo_version = "1"
}
