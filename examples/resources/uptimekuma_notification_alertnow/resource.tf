resource "uptimekuma_notification_alertnow" "example" {
  name        = "AlertNow Notifications"
  webhook_url = "https://alertnow.example.com/webhook/your-token"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_alertnow" "backup" {
  name        = "AlertNow Backup"
  webhook_url = "https://alertnow.example.com/webhook/backup-token"
  is_active   = true
  is_default  = false
}
