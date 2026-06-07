resource "uptimekuma_notification_halopsa" "example" {
  name        = "HaloPSA Notifications"
  webhook_url = "https://halopsa.example.com/webhook/YOUR_TOKEN"
  username    = "your-username"
  password    = "your-password"
  is_active   = true
  is_default  = false
}
