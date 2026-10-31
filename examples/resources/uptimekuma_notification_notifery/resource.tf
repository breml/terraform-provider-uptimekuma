resource "uptimekuma_notification_notifery" "example" {
  name       = "Notifery Notifications"
  api_key    = "your-notifery-api-key"
  title      = "Uptime Kuma Alert"
  is_default = false
}

resource "uptimekuma_notification_notifery" "with_group" {
  name       = "Notifery with Group"
  api_key    = "your-notifery-api-key"
  title      = "Uptime Kuma Alert"
  group      = "infrastructure"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in plan
# or state. Change the matching *_wo_version whenever a secret changes.
resource "uptimekuma_notification_notifery" "write_only" {
  name = "Notifery with write-only secret"

  api_key_wo         = "your-notifery-api-key"
  api_key_wo_version = "1"
}
