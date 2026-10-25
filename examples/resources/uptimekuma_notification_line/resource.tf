resource "uptimekuma_notification_line" "example" {
  name                 = "LINE Notification"
  channel_access_token = "your_channel_access_token"
  user_id              = "your_user_id"
  is_active            = true
}

# Requires Terraform or OpenTofu 1.11 or later: channel_access_token_wo is never
# stored in plan or state. Change channel_access_token_wo_version whenever the
# secret changes.
resource "uptimekuma_notification_line" "write_only" {
  name    = "LINE with write-only token"
  user_id = "your_user_id"

  channel_access_token_wo         = "your_channel_access_token"
  channel_access_token_wo_version = "1"
}
