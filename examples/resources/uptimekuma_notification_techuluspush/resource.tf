resource "uptimekuma_notification_techuluspush" "example" {
  name       = "TechulusPush Notifications"
  api_key    = "your-techulus-push-api-key"
  title      = "Uptime Kuma Alert"
  channel    = "monitoring"
  is_default = false
}

resource "uptimekuma_notification_techuluspush" "urgent" {
  name           = "TechulusPush Urgent Alerts"
  api_key        = "your-techulus-push-api-key"
  title          = "URGENT: Uptime Kuma"
  sound          = "alarm"
  time_sensitive = true
  is_default     = true
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_techuluspush" "write_only" {
  name = "TechulusPush with a write-only secret"

  api_key_wo         = "your-techulus-push-api-key"
  api_key_wo_version = "1"
}
