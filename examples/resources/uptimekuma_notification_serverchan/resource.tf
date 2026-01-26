resource "uptimekuma_notification_serverchan" "example" {
  name       = "ServerChan Notification"
  is_active  = true
  is_default = false

  # ServerChan send key for authentication
  # Generate this from your ServerChan account
  send_key = "SCT12345abcdef"
}
