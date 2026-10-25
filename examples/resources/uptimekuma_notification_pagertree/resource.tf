resource "uptimekuma_notification_pagertree" "example" {
  name            = "PagerTree Notification"
  is_active       = true
  integration_url = "https://alerts.pagertree.com/api/v2/incidents"
  urgency         = "high"
  auto_resolve    = "resolve"
  apply_existing  = false
}

# Requires Terraform or OpenTofu 1.11 or later: integration_url_wo is never stored in
# plan or state. Change integration_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_pagertree" "write_only" {
  name = "PagerTree with a write-only secret"

  integration_url_wo         = "https://api.pagertree.com/integration/your-integration-id"
  integration_url_wo_version = "1"
}
