# PushDeer notification resource with default server
resource "uptimekuma_notification_pushdeer" "example" {
  name      = "PushDeer Notification"
  key       = "pushkey123"
  is_active = true
}

# PushDeer notification resource with custom server
resource "uptimekuma_notification_pushdeer" "custom_server" {
  name      = "PushDeer Custom Server"
  key       = "pushkey456"
  server    = "https://custom.pushdeer.server.com"
  is_active = true
}

# Requires Terraform or OpenTofu 1.11 or later: key_wo is never stored in
# plan or state. Change key_wo_version whenever the secret changes.
resource "uptimekuma_notification_pushdeer" "write_only" {
  name = "PushDeer with a write-only secret"

  key_wo         = "your-pushdeer-push-key"
  key_wo_version = "1"
}
