resource "uptimekuma_notification_zohocliq" "example" {
  name        = "Zoho Cliq Notifications"
  webhook_url = "https://cliq.zoho.com/company/api/v2/channelsbyname/general/message?zapikey=your-api-key"
  is_default  = false
}

resource "uptimekuma_notification_zohocliq" "backup" {
  name        = "Zoho Cliq Backup"
  webhook_url = "https://cliq.zoho.com/company/api/v2/channelsbyname/alerts/message?zapikey=your-backup-key"
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_zohocliq" "write_only" {
  name = "Zoho Cliq with a write-only secret"

  webhook_url_wo         = "https://cliq.zoho.com/company/api/v2/channelsbyname/general/message?zapikey=your-api-key"
  webhook_url_wo_version = "1"
}
