resource "uptimekuma_notification_flashduty" "example" {
  name            = "FlashDuty Notifications"
  integration_key = "https://api.flashduty.com/webhook/events/YOUR_INTEGRATION_KEY"
  severity        = "Critical"
  is_active       = true
  is_default      = false
}

resource "uptimekuma_notification_flashduty" "warning_alerts" {
  name            = "FlashDuty Warning Alerts"
  integration_key = "https://api.flashduty.com/webhook/events/YOUR_INTEGRATION_KEY"
  severity        = "Warning"
  is_active       = true
  is_default      = false
}
