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
