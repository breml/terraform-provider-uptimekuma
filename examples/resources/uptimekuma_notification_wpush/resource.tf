resource "uptimekuma_notification_wpush" "example" {
  name      = "WPush Notifications"
  api_key   = "your-wpush-api-key"
  channel   = "monitoring"
  is_active = true
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_wpush" "write_only" {
  name = "WPush with a write-only secret"

  channel = "monitoring"

  api_key_wo         = "your-wpush-api-key"
  api_key_wo_version = "1"
}
