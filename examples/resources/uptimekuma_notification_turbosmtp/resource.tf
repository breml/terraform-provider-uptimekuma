resource "uptimekuma_notification_turbosmtp" "example" {
  name            = "TurboSMTP Notifications"
  consumer_key    = "your-turbosmtp-consumer-key"
  consumer_secret = "your-turbosmtp-consumer-secret"
  from_email      = "uptime-kuma@example.com"
  to_email        = "oncall@example.com"
  is_default      = false
}

# Send through the EU API host, with CC and BCC recipients and a custom subject.
resource "uptimekuma_notification_turbosmtp" "eu" {
  name            = "TurboSMTP EU"
  consumer_key    = "your-turbosmtp-consumer-key"
  consumer_secret = "your-turbosmtp-consumer-secret"
  region          = "eu"
  from_email      = "uptime-kuma@example.eu"
  to_email        = "oncall@example.eu,ops@example.eu"
  cc_email        = "team-lead@example.eu"
  bcc_email       = "audit@example.eu"
  subject         = "Uptime Kuma alert"
  is_default      = false
}

# Requires Terraform or OpenTofu 1.11 or later: consumer_secret_wo is never stored in
# plan or state. Change consumer_secret_wo_version whenever the secret changes.
resource "uptimekuma_notification_turbosmtp" "write_only" {
  name = "TurboSMTP with a write-only secret"

  consumer_key = "your-turbosmtp-consumer-key"
  from_email   = "uptime-kuma@example.com"
  to_email     = "oncall@example.com"

  consumer_secret_wo         = "your-turbosmtp-consumer-secret"
  consumer_secret_wo_version = "1"
}
