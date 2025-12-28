resource "uptimekuma_notification_rocketchat" "example" {
  name        = "RocketChat Notifications"
  webhook_url = "https://rocket.example.com/hooks/uid/token"
  channel     = "general"
  icon_emoji  = ":bell:"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_rocketchat" "with_button" {
  name        = "RocketChat with Button"
  webhook_url = "https://rocket.example.com/hooks/uid/token"
  username    = "Uptime Kuma"
  icon_emoji  = ":warning:"
  channel     = "alerts"
  button      = "View Status"
  is_active   = true
  is_default  = false
}
