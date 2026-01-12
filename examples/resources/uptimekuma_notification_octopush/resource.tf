# Octopush V2 API Example
resource "uptimekuma_notification_octopush" "example_v2" {
  name = "Octopush SMS V2"

  # API Version (1 for Direct Mail, 2 for standard API)
  version = "2"

  # V2 API Configuration
  api_key      = "your-api-key"
  login        = "your-login"
  phone_number = "+1234567890"
  sms_type     = "sms_premium" # or "sms_low_cost"
  sender_name  = "MyApp"

  # Optional: Enable or disable the notification
  is_active = true

  # Optional: Mark as default notification
  is_default = false

  # Optional: Apply to existing monitors
  apply_existing = false
}

# Octopush V1 (Direct Mail) API Example
resource "uptimekuma_notification_octopush" "example_v1" {
  name = "Octopush Direct Mail V1"

  # API Version
  version = "1"

  # V1 (Direct Mail) API Configuration
  dm_api_key      = "your-dm-api-key"
  dm_login        = "your-dm-login"
  dm_phone_number = "+1234567890"
  dm_sms_type     = "your-sms-type"
  dm_sender_name  = "MyApp"

  is_active = true
}
