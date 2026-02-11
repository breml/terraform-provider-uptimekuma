resource "uptimekuma_notification_heiioncall" "example" {
  name       = "Heii On-Call Alerts"
  api_key    = "your-heii-oncall-api-key"
  trigger_id = "your-heii-oncall-trigger-id"
  is_active  = true
  is_default = false
}
