resource "uptimekuma_notification_keep" "example" {
  name        = "Keep Notifications"
  webhook_url = "https://api.keephq.dev/alerts/alert"
  api_key     = "your-keep-api-key"
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo and api_key_wo are
# never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_keep" "write_only" {
  name = "Keep with write-only secrets"

  webhook_url_wo         = "https://api.keephq.dev/alerts/alert"
  webhook_url_wo_version = "1"
  api_key_wo             = "your-keep-api-key"
  api_key_wo_version     = "1"
}
