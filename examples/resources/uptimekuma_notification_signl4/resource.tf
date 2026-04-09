resource "uptimekuma_notification_signl4" "example" {
  name        = "SIGNL4 Alerts"
  webhook_url = "https://connect.signl4.com/webhook/YOUR_TEAM_SECRET"
  is_active   = true
  is_default  = false
}
