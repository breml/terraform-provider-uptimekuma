resource "uptimekuma_notification_grafanaoncall" "example" {
  name               = "Grafana OnCall Notifications"
  is_default         = false
  apply_existing     = false
  grafana_oncall_url = "https://grafana-oncall.example.com/integrations/v1/webhook/abc123/"
}

# Requires Terraform or OpenTofu 1.11 or later: grafana_oncall_url_wo is never stored in
# plan or state. Change grafana_oncall_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_grafanaoncall" "write_only" {
  name = "Grafana OnCall with a write-only secret"

  grafana_oncall_url_wo         = "https://oncall.example.com/integrations/v1/formatted_webhook/your-token/"
  grafana_oncall_url_wo_version = "1"
}
