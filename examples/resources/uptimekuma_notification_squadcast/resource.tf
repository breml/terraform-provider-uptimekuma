resource "uptimekuma_notification_squadcast" "example" {
  name        = "Squadcast Alerts"
  webhook_url = "https://api.squadcast.com/v3/incidents/webhook/YOUR_WEBHOOK_KEY"
  is_active   = true
  is_default  = false
}
