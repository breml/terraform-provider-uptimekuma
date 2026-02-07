resource "uptimekuma_notification_goalert" "example" {
  name      = "Example GoAlert Notification"
  is_active = true
  base_url  = "https://goalert.example.com"
  token     = "your-integration-token"
}
