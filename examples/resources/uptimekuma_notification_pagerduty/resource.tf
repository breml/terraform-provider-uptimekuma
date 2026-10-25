resource "uptimekuma_notification_pagerduty" "example" {
  name            = "Production Alerts"
  is_active       = true
  integration_url = "https://events.pagerduty.com/integration/YOUR_INTEGRATION_ID/enqueue"
  integration_key = "YOUR_PAGERDUTY_KEY"
  priority        = "high"
  auto_resolve    = "resolved"
}

# Requires Terraform or OpenTofu 1.11 or later: integration_url_wo and integration_key_wo
# are never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_pagerduty" "write_only" {
  name = "PagerDuty with write-only secrets"

  integration_url_wo         = "https://events.pagerduty.com/integration/YOUR_INTEGRATION_ID/enqueue"
  integration_url_wo_version = "1"
  integration_key_wo         = "your-pagerduty-integration-key"
  integration_key_wo_version = "1"
}
