resource "uptimekuma_notification_googlesheets" "example" {
  name        = "Google Sheets Notifications"
  webhook_url = "https://script.google.com/macros/s/AKfy.../exec"
  is_active   = true
  is_default  = false
}
