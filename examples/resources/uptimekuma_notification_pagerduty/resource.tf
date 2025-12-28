resource "uptimekuma_notification_pagerduty" "example" {
  name            = "Production Alerts"
  is_active       = true
  integration_url = "https://events.pagerduty.com/integration/YOUR_INTEGRATION_ID/enqueue"
  integration_key = "YOUR_PAGERDUTY_KEY"
  priority        = "high"
  auto_resolve    = "resolved"
}
