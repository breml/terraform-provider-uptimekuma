resource "uptimekuma_notification_goalert" "example" {
  name      = "Example GoAlert Notification"
  is_active = true
  base_url  = "https://goalert.example.com"
  token     = "your-integration-token"
}

# Requires Terraform or OpenTofu 1.11 or later: token_wo is never stored in
# plan or state. Change token_wo_version whenever the secret changes.
resource "uptimekuma_notification_goalert" "write_only" {
  name = "GoAlert with a write-only secret"

  base_url = "https://goalert.example.com"

  token_wo         = "your-goalert-token"
  token_wo_version = "1"
}
