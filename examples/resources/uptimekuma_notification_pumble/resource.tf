resource "uptimekuma_notification_pumble" "example" {
  name        = "Pumble Notifications"
  webhook_url = "https://api.pumble.com/workspaces/YOUR_WORKSPACE_ID/incomingWebhooks/YOUR_WEBHOOK_ID"
  is_active   = true
  is_default  = false
}
