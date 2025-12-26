resource "uptimekuma_notification_discord" "example" {
  name        = "Discord Notifications"
  webhook_url = "https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz"
  username    = "Uptime Kuma"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_discord" "alerts_channel" {
  name           = "Discord Critical Alerts"
  webhook_url    = "https://discord.com/api/webhooks/987654321/zyxwvutsrqponmlkjihgfedcba"
  username       = "Uptime Kuma Alerts"
  prefix_message = "🚨 ALERT"
  disable_url    = false
  is_active      = true
  is_default     = true
}
