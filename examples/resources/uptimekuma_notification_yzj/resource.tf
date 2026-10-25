resource "uptimekuma_notification_yzj" "example" {
  name        = "YZJ Notifications"
  webhook_url = "https://yzj.example.com/webhook/your-webhook"
  token       = "your-yzj-token"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_yzj" "alerts" {
  name        = "YZJ Alerts"
  webhook_url = "https://yzj.example.com/webhook/alerts"
  token       = "your-yzj-alert-token"
  is_active   = true
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo and token_wo are
# never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_yzj" "write_only" {
  name = "YZJ with write-only secrets"

  webhook_url_wo         = "https://yzj.example.com/webhook/your-webhook"
  webhook_url_wo_version = "1"
  token_wo               = "your-yzj-token"
  token_wo_version       = "1"
}
