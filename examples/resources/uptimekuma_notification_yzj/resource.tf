resource "uptimekuma_notification_yzj" "example" {
  name        = "YZJ Notifications"
  webhook_url = "https://yzj.example.com/webhook/your-webhook"
  token       = "your-yzj-token"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_yzj" "alerts" {
  name        = "YZJ Alerts"
  webhook_url = "https://yzj.example.com/webhook/alerts"
  token       = "your-yzj-alert-token"
  is_active   = true
  is_default  = false
}
