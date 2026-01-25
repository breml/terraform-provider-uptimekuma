resource "uptimekuma_notification_sendgrid" "example" {
  name       = "SendGrid Notifications"
  api_key    = "SG.your-api-key-here"
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  subject    = "Uptime Alert Notification"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_sendgrid" "with_cc_bcc" {
  name       = "SendGrid with CC/BCC"
  api_key    = "SG.your-api-key-here"
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  subject    = "Critical Alert"
  cc_email   = "team-lead@example.com"
  bcc_email  = "archive@example.com"
  is_active  = true
  is_default = false
}
