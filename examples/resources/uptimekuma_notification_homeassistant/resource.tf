resource "uptimekuma_notification_homeassistant" "example" {
  name                    = "Home Assistant Notifications"
  is_active               = true
  is_default              = false
  apply_existing          = false
  home_assistant_url      = "https://homeassistant.example.com"
  long_lived_access_token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
  notification_service    = "notify.mobile_app"
}

# Requires Terraform or OpenTofu 1.11 or later: long_lived_access_token_wo is
# never stored in plan or state. Change long_lived_access_token_wo_version
# whenever the secret changes.
resource "uptimekuma_notification_homeassistant" "write_only" {
  name                 = "Home Assistant with write-only token"
  home_assistant_url   = "https://homeassistant.example.com"
  notification_service = "notify.mobile_app"

  long_lived_access_token_wo         = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
  long_lived_access_token_wo_version = "1"
}
