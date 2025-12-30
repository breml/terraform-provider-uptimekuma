resource "uptimekuma_notification_line" "example" {
  name                 = "LINE Notification"
  channel_access_token = "your_channel_access_token"
  user_id              = "your_user_id"
  is_active            = true
}
