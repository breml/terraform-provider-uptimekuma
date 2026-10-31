resource "uptimekuma_notification_pushy" "example" {
  name       = "My Pushy Notification"
  is_default = false

  # Pushy API secret key (sensitive)
  api_key = "your-pushy-api-key"

  # Device token for push notifications (sensitive)
  token = "your-device-token"

  # Optional: apply this notification to all existing monitors
  apply_existing = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_pushy" "write_only" {
  name = "Pushy with a write-only secret"

  token = "your-device-token"

  api_key_wo         = "your-pushy-api-key"
  api_key_wo_version = "1"
}
