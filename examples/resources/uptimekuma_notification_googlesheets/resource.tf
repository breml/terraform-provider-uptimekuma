resource "uptimekuma_notification_googlesheets" "example" {
  name        = "Google Sheets Notifications"
  webhook_url = "https://script.google.com/macros/s/AKfy.../exec"
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_googlesheets" "write_only" {
  name = "Google Sheets with a write-only secret"

  webhook_url_wo         = "https://script.google.com/macros/s/your-deployment-id/exec"
  webhook_url_wo_version = "1"
}
