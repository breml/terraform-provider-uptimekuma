resource "uptimekuma_notification_apprise" "example" {
  name        = "Apprise Notifications"
  apprise_url = "discord://webhook_id/webhook_token"
  title       = "Uptime Kuma Alert"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_apprise" "multi_service" {
  name        = "Multi-Service Apprise"
  apprise_url = "discord://webhook1/token1,slack://tokenA/tokenB/tokenC"
  title       = "Critical System Alert"
  is_active   = true
  is_default  = true
}
