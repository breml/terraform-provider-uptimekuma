resource "uptimekuma_notification_pushover" "example" {
  name      = "My Pushover Notification"
  is_active = true
  user_key  = var.pushover_user_key
  app_token = var.pushover_app_token
}
