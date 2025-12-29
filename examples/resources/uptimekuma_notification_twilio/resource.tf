resource "uptimekuma_notification_twilio" "example" {
  name           = "Twilio SMS Notifications"
  account_sid    = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  auth_token     = "your-auth-token"
  to_number      = "+12025550123"
  from_number    = "+12025550456"
  is_active      = true
  is_default     = false
}

resource "uptimekuma_notification_twilio" "with_api_key" {
  name           = "Twilio with API Key"
  account_sid    = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  api_key        = "your-api-key"
  auth_token     = "your-auth-token"
  to_number      = "+12025550123"
  from_number    = "+12025550456"
  is_active      = true
  is_default     = false
}
