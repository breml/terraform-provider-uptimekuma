resource "uptimekuma_notification_discord" "example" {
  name        = "Discord Notifications"
  webhook_url = "https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz"
  username    = "Uptime Kuma"
  is_default  = false
}

resource "uptimekuma_notification_discord" "alerts_channel" {
  name           = "Discord Critical Alerts"
  webhook_url    = "https://discord.com/api/webhooks/987654321/zyxwvutsrqponmlkjihgfedcba"
  username       = "Uptime Kuma Alerts"
  prefix_message = "🚨 ALERT"
  disable_url    = false
  is_default     = true
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_discord" "write_only" {
  name = "Discord with a write-only secret"

  webhook_url_wo         = "https://discord.com/api/webhooks/123/your-token"
  webhook_url_wo_version = "1"
}
