resource "uptimekuma_notification_googlechat" "example" {
  name        = "My Google Chat Notification"
  webhook_url = "https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=KEY&token=TOKEN"
  is_active   = true

  # Optional: Use a custom message template
  use_template = true
  template     = "Alert: {msg}"
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_googlechat" "write_only" {
  name = "Google Chat with a write-only secret"

  webhook_url_wo         = "https://chat.googleapis.com/v1/spaces/AAAA/messages?key=your-key&token=your-token"
  webhook_url_wo_version = "1"
}
