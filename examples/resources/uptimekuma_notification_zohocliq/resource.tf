resource "uptimekuma_notification_zohocliq" "example" {
  name        = "Zoho Cliq Notifications"
  webhook_url = "https://cliq.zoho.com/company/api/v2/channelsbyname/general/message?zapikey=your-api-key"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_zohocliq" "backup" {
  name        = "Zoho Cliq Backup"
  webhook_url = "https://cliq.zoho.com/company/api/v2/channelsbyname/alerts/message?zapikey=your-backup-key"
  is_active   = true
  is_default  = false
}
