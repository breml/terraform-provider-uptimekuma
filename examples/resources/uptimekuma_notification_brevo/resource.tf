resource "uptimekuma_notification_brevo" "example" {
  name       = "Brevo Email Alerts"
  is_active  = true
  api_key    = "your_brevo_api_key_here"
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  from_name  = "Uptime Monitoring"
  subject    = "Uptime Alert Notification"
}

resource "uptimekuma_notification_brevo" "with_cc" {
  name       = "Brevo Email Alerts with CC"
  is_active  = true
  api_key    = "your_brevo_api_key_here"
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  from_name  = "Uptime Monitoring"
  subject    = "Uptime Alert Notification"
  cc_email   = "team@example.com,manager@example.com"
  bcc_email  = "archive@example.com"
}
