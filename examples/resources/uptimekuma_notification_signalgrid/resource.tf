resource "uptimekuma_notification_signalgrid" "example" {
  name       = "Signalgrid Notifications"
  client_key = "your-signalgrid-client-key"
  channel    = "alerts"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: client_key_wo is never stored in
# plan or state. Change client_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_signalgrid" "write_only" {
  name    = "Signalgrid with a write-only secret"
  channel = "alerts"

  client_key_wo         = "your-signalgrid-client-key"
  client_key_wo_version = "1"
}
