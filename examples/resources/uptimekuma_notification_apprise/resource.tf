resource "uptimekuma_notification_apprise" "example" {
  name        = "Apprise Notifications"
  apprise_url = "discord://webhook_id/webhook_token"
  title       = "Uptime Kuma Alert"
  is_default  = false
}

resource "uptimekuma_notification_apprise" "multi_service" {
  name        = "Multi-Service Apprise"
  apprise_url = "discord://webhook1/token1,slack://tokenA/tokenB/tokenC"
  title       = "Critical System Alert"
  is_default  = true
}

# Requires Terraform or OpenTofu 1.11 or later: apprise_url_wo is never stored in
# plan or state. Change apprise_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_apprise" "write_only" {
  name = "Apprise with a write-only secret"

  apprise_url_wo         = "tgram://bot-token/chat-id"
  apprise_url_wo_version = "1"
}
