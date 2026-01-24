resource "uptimekuma_notification_pushy" "example" {
  name       = "My Pushy Notification"
  is_active  = true
  is_default = false

  # Pushy API secret key (sensitive)
  api_key = "your-pushy-api-key"

  # Device token for push notifications (sensitive)
  token = "your-device-token"

  # Optional: apply this notification to all existing monitors
  apply_existing = false
}
