resource "uptimekuma_notification_grafanaoncall" "example" {
  name               = "Grafana OnCall Notifications"
  is_active          = true
  is_default         = false
  apply_existing     = false
  grafana_oncall_url = "https://grafana-oncall.example.com/integrations/v1/webhook/abc123/"
}
