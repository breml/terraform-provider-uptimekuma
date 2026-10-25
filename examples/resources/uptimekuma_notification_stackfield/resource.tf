resource "uptimekuma_notification_stackfield" "example" {
  name        = "Stackfield Notifications"
  webhook_url = "https://api.stackfield.com/hooks/YOUR_WEBHOOK_ID"
  is_active   = true
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_stackfield" "write_only" {
  name = "Stackfield with a write-only secret"

  webhook_url_wo         = "https://api.stackfield.com/hooks/YOUR_WEBHOOK_ID"
  webhook_url_wo_version = "1"
}
