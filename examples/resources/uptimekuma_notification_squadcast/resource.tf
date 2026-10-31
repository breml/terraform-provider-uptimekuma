resource "uptimekuma_notification_squadcast" "example" {
  name        = "Squadcast Alerts"
  webhook_url = "https://api.squadcast.com/v3/incidents/webhook/YOUR_WEBHOOK_KEY"
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_squadcast" "write_only" {
  name = "Squadcast with a write-only secret"

  webhook_url_wo         = "https://api.squadcast.com/v3/incidents/webhook/YOUR_WEBHOOK_KEY"
  webhook_url_wo_version = "1"
}
