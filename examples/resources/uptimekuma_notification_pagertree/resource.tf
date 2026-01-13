resource "uptimekuma_notification_pagertree" "example" {
  name            = "PagerTree Notification"
  is_active       = true
  integration_url = "https://alerts.pagertree.com/api/v2/incidents"
  urgency         = "high"
  auto_resolve    = "resolve"
  apply_existing  = false
}
