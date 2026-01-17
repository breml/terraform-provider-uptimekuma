# Basic Nextcloud Talk notification
resource "uptimekuma_notification_nextcloudtalk" "example" {
  name               = "Nextcloud Talk Notifications"
  host               = "https://nextcloud.example.com"
  conversation_token = "abc123token"
  bot_secret         = "your-bot-secret"
  is_active          = true
  is_default         = false
}

# Nextcloud Talk notification with silent options
resource "uptimekuma_notification_nextcloudtalk" "silent" {
  name               = "Silent Nextcloud Alerts"
  host               = "https://cloud.example.org"
  conversation_token = "room-token-xyz"
  bot_secret         = "another-bot-secret"
  send_silent_up     = true
  send_silent_down   = false
  is_active          = true
  is_default         = false
}
