resource "uptimekuma_notification_stackfield" "example" {
  name        = "Stackfield Notifications"
  webhook_url = "https://api.stackfield.com/hooks/YOUR_WEBHOOK_ID"
  is_active   = true
  is_default  = false
}
