resource "uptimekuma_notification_pushover" "example" {
  name      = "My Pushover Notification"
  user_key  = var.pushover_user_key
  app_token = var.pushover_app_token
}

# Requires Terraform or OpenTofu 1.11 or later: app_token_wo is never stored in
# plan or state. Change app_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_pushover" "write_only" {
  name = "Pushover with a write-only secret"

  user_key = "your-pushover-user-key"

  app_token_wo         = "your-pushover-app-token"
  app_token_wo_version = "1"
}
