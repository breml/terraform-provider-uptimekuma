resource "uptimekuma_notification_mattermost" "example" {
  name        = "Mattermost Example"
  is_active   = true
  webhook_url = "https://mattermost.example.com/hooks/xxx"
  username    = "Uptime Kuma"
  channel     = "#alerts"
  icon_emoji  = ":robot_face:"
  icon_url    = "https://example.com/icon.png"
}
