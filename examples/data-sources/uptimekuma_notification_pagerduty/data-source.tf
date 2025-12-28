data "uptimekuma_notification_pagerduty" "example" {
  name = "Production Alerts"
}

# Or retrieve by ID
data "uptimekuma_notification_pagerduty" "by_id" {
  id = 1
}
