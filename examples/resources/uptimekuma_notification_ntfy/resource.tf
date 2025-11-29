resource "uptimekuma_notification_ntfy" "example" {
  name       = "ntfy.sh Notifications"
  topic      = "my_uptime_kuma_notifications"
  priority   = 5
  server_url = "https://ntfy.sh"
  icon       = "bell"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_ntfy" "high_priority" {
  name       = "ntfy Critical Alerts"
  topic      = "critical_alerts"
  priority   = 1
  server_url = "https://ntfy.sh"
  is_active  = true
  is_default = true
}
