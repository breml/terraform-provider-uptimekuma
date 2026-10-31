resource "uptimekuma_notification_brevo" "example" {
  name       = "Brevo Email Alerts"
  api_key    = "your_brevo_api_key_here"
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  from_name  = "Uptime Monitoring"
  subject    = "Uptime Alert Notification"
}

resource "uptimekuma_notification_brevo" "with_cc" {
  name       = "Brevo Email Alerts with CC"
  api_key    = "your_brevo_api_key_here"
  to_email   = "alerts@example.com"
  from_email = "monitoring@example.com"
  from_name  = "Uptime Monitoring"
  subject    = "Uptime Alert Notification"
  cc_email   = "team@example.com,manager@example.com"
  bcc_email  = "archive@example.com"
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_brevo" "write_only" {
  name = "Brevo with a write-only secret"

  to_email   = "ops@example.com"
  from_email = "uptime@example.com"

  api_key_wo         = "your-brevo-api-key"
  api_key_wo_version = "1"
}
