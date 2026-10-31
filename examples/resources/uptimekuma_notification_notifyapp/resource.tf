resource "uptimekuma_notification_notifyapp" "example" {
  name       = "Notify! Notifications"
  device_id  = "ABC12345"
  token      = "your-notify-token"
  icon_url   = "https://example.com/icon.png"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: token_wo is never stored in
# plan or state. Change token_wo_version whenever the secret changes.
resource "uptimekuma_notification_notifyapp" "write_only" {
  name      = "Notify! with a write-only secret"
  device_id = "ABC12345"

  token_wo         = "your-notify-token"
  token_wo_version = "1"
}
