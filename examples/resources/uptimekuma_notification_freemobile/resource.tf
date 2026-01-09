resource "uptimekuma_notification_freemobile" "example" {
  name       = "Free Mobile SMS Alerts"
  user       = "YOUR_FREE_MOBILE_USER_ID"
  pass       = "YOUR_FREE_MOBILE_API_KEY"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_freemobile" "production" {
  name       = "Free Mobile Production Alerts"
  user       = "1234567890"
  pass       = "your_api_key_here"
  is_active  = true
  is_default = true
}
