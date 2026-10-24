resource "uptimekuma_notification_clickup" "example" {
  name         = "ClickUp Notifications"
  token        = "pk_12345678_ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"
  workspace_id = "9012345678"
  channel_id   = "abc12-345"
  is_active    = true
  is_default   = false
}

# Leave the monitor address out of the message body
resource "uptimekuma_notification_clickup" "no_url" {
  name         = "ClickUp Notifications without URL"
  token        = "pk_12345678_ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"
  workspace_id = "9012345678"
  channel_id   = "abc12-345"
  disable_url  = true
}
