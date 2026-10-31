resource "uptimekuma_notification_gotify" "example" {
  name              = "Gotify Notifications"
  is_default        = false
  apply_existing    = false
  server_url        = "https://gotify.example.com"
  application_token = "AGe0Ks4WV5fEJkX"
  priority          = 8
}

# Requires Terraform or OpenTofu 1.11 or later: application_token_wo is never stored in
# plan or state. Change application_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_gotify" "write_only" {
  name = "Gotify with a write-only secret"

  server_url = "https://gotify.example.com"

  application_token_wo         = "your-gotify-app-token"
  application_token_wo_version = "1"
}
