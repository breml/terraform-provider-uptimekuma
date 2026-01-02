resource "uptimekuma_notification_bitrix24" "example" {
  name                 = "Example Bitrix24 Notification"
  is_active            = true
  webhook_url          = "https://your-bitrix24-domain.bitrix24.com/rest/1/webhook-key"
  notification_user_id = "123"
}
