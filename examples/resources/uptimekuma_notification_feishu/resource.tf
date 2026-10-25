resource "uptimekuma_notification_feishu" "example" {
  name        = "Feishu Notifications"
  webhook_url = "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_feishu" "critical" {
  name        = "Feishu Critical Alerts"
  webhook_url = "https://open.feishu.cn/open-apis/bot/v2/hook/critical-alerts-webhook"
  is_active   = true
  is_default  = true
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_feishu" "write_only" {
  name = "Feishu with a write-only secret"

  webhook_url_wo         = "https://open.feishu.cn/open-apis/bot/v2/hook/your-token"
  webhook_url_wo_version = "1"
}
