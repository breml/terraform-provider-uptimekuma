resource "uptimekuma_notification_linenotify" "example" {
  name         = "LINE Notify Alert"
  is_active    = true
  access_token = "YOUR_LINE_NOTIFY_ACCESS_TOKEN"

  # Optional fields
  is_default     = false
  apply_existing = false
}
