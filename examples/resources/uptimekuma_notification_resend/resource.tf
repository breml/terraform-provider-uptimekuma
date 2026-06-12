resource "uptimekuma_notification_resend" "example" {
  name       = "Resend Notifications"
  api_key    = "re_your_api_key_here"
  from_email = "monitoring@example.com"
  to_email   = "alerts@example.com"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_resend" "with_subject" {
  name       = "Resend with Custom Subject"
  api_key    = "re_your_api_key_here"
  from_email = "monitoring@example.com"
  from_name  = "Uptime Kuma"
  to_email   = "alerts@example.com,oncall@example.com"
  subject    = "Uptime Alert Notification"
  is_active  = true
  is_default = false
}
