resource "uptimekuma_notification_teams" "example" {
  name        = "Microsoft Teams Notifications"
  webhook_url = "https://outlook.webhook.office.com/webhookb2/..."
  is_active   = true
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_teams" "write_only" {
  name = "Microsoft Teams with a write-only secret"

  webhook_url_wo         = "https://outlook.webhook.office.com/webhookb2/..."
  webhook_url_wo_version = "1"
}
