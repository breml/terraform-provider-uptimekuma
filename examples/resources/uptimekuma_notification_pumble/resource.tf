resource "uptimekuma_notification_pumble" "example" {
  name        = "Pumble Notifications"
  webhook_url = "https://api.pumble.com/workspaces/YOUR_WORKSPACE_ID/incomingWebhooks/YOUR_WEBHOOK_ID"
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_pumble" "write_only" {
  name = "Pumble with a write-only secret"

  webhook_url_wo         = "https://api.pumble.com/workspaces/YOUR_WORKSPACE_ID/incomingWebhooks/YOUR_WEBHOOK_ID"
  webhook_url_wo_version = "1"
}
