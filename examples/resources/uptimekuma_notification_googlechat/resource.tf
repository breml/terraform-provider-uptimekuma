resource "uptimekuma_notification_googlechat" "example" {
  name        = "My Google Chat Notification"
  webhook_url = "https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=KEY&token=TOKEN"
  is_active   = true

  # Optional: Use a custom message template
  use_template = true
  template     = "Alert: {msg}"
}
