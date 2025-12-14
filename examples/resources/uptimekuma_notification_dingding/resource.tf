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
