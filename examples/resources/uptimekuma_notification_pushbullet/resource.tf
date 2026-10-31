resource "uptimekuma_notification_pushbullet" "example" {
  name         = "My Pushbullet Notification"
  access_token = var.pushbullet_access_token
}

# Requires Terraform or OpenTofu 1.11 or later: access_token_wo is never stored in
# plan or state. Change access_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_pushbullet" "write_only" {
  name = "Pushbullet with a write-only secret"

  access_token_wo         = "your-pushbullet-access-token"
  access_token_wo_version = "1"
}
