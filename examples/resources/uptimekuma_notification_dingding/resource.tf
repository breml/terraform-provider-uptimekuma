resource "uptimekuma_notification_dingding" "example" {
  name        = "DingDing Notifications"
  webhook_url = "https://oapi.dingtalk.com/robot/send?access_token=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  secret_key  = "SECxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_dingding" "critical" {
  name        = "DingDing Critical Alerts"
  webhook_url = "https://oapi.dingtalk.com/robot/send?access_token=critical-token"
  secret_key  = "SECcritical-secret-key"
  mentioning  = "@all"
  is_active   = true
  is_default  = true
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo and secret_key_wo are
# never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_dingding" "write_only" {
  name = "DingDing with write-only secrets"

  webhook_url_wo         = "https://oapi.dingtalk.com/robot/send?access_token=your-token"
  webhook_url_wo_version = "1"
  secret_key_wo          = "your-signing-secret"
  secret_key_wo_version  = "1"
}
