resource "uptimekuma_notification_pushbullet" "example" {
  name         = "My Pushbullet Notification"
  is_active    = true
  access_token = var.pushbullet_access_token
}
