resource "uptimekuma_notification_opsgenie" "example" {
  name      = "OpsGenie Example"
  is_active = true
  api_key   = "your-opsgenie-api-key"
  region    = "us"
  priority  = 2
}
