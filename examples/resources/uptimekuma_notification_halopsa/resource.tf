resource "uptimekuma_notification_halopsa" "example" {
  name        = "HaloPSA Notifications"
  webhook_url = "https://halopsa.example.com/webhook/YOUR_TOKEN"
  username    = "your-username"
  password    = "your-password"
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo and password_wo are
# never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_halopsa" "write_only" {
  name     = "HaloPSA with write-only secrets"
  username = "your-username"

  webhook_url_wo         = "https://halopsa.example.com/webhook/YOUR_TOKEN"
  webhook_url_wo_version = "1"
  password_wo            = "your-password"
  password_wo_version    = "1"
}
