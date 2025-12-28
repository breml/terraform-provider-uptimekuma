resource "uptimekuma_notification_homeassistant" "example" {
  name                    = "Home Assistant Notifications"
  is_active               = true
  is_default              = false
  apply_existing          = false
  home_assistant_url      = "https://homeassistant.example.com"
  long_lived_access_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
  notification_service    = "notify.mobile_app"
}
