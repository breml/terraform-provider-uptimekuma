resource "uptimekuma_notification_bitrix24" "example" {
  name                 = "Example Bitrix24 Notification"
  webhook_url          = "https://your-bitrix24-domain.bitrix24.com/rest/1/webhook-key"
  notification_user_id = "123"
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_bitrix24" "write_only" {
  name = "Bitrix24 with a write-only secret"

  notification_user_id = "1"

  webhook_url_wo         = "https://example.bitrix24.com/rest/1/your-webhook-key/"
  webhook_url_wo_version = "1"
}
