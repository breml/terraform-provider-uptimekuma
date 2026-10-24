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

# Requires Terraform or OpenTofu 1.11 or later: integration_key_wo is never stored in
# plan or state. Change integration_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_flashduty" "write_only" {
  name = "FlashDuty with a write-only secret"

  integration_key_wo         = "your-integration-key"
  integration_key_wo_version = "1"
}
