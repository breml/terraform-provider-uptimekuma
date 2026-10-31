resource "uptimekuma_notification_sendgrid" "example" {
  name       = "SendGrid Notifications"
  api_key    = "SG.your-api-key-here"
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  subject    = "Uptime Alert Notification"
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
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_sendgrid" "write_only" {
  name = "SendGrid with a write-only secret"

  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  subject    = "Uptime Alert Notification"

  api_key_wo         = "SG.your-api-key-here"
  api_key_wo_version = "1"
}
