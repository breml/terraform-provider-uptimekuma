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

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_alertnow" "write_only" {
  name = "AlertNow with a write-only secret"

  webhook_url_wo         = "https://alertnow.example.com/webhook/your-key"
  webhook_url_wo_version = "1"
}
