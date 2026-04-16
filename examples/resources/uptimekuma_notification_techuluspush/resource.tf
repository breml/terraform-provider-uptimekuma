resource "uptimekuma_notification_techuluspush" "example" {
  name       = "TechulusPush Notifications"
  api_key    = "your-techulus-push-api-key"
  title      = "Uptime Kuma Alert"
  channel    = "monitoring"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_techuluspush" "urgent" {
  name           = "TechulusPush Urgent Alerts"
  api_key        = "your-techulus-push-api-key"
  title          = "URGENT: Uptime Kuma"
  sound          = "alarm"
  time_sensitive = true
  is_active      = true
  is_default     = true
}
