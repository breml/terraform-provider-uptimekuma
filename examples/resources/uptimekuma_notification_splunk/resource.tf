resource "uptimekuma_notification_splunk" "example" {
  name            = "Splunk Notification"
  rest_url        = "https://api.victorops.com/api/v2"
  severity        = "critical"
  auto_resolve    = "resolve"
  integration_key = "your_integration_key"
  is_active       = true
}
