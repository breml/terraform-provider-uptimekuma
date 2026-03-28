resource "uptimekuma_notification_sevenio" "example" {
  name      = "Seven.io SMS Notification"
  is_active = true
  api_key   = "your_sevenio_api_key"
  sender    = "+491234567890"
  to        = "+491111111111"
}
