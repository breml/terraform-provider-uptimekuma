resource "uptimekuma_notification_teams" "example" {
  name        = "Microsoft Teams Notifications"
  webhook_url = "https://outlook.webhook.office.com/webhookb2/..."
  is_active   = true
  is_default  = false
}
