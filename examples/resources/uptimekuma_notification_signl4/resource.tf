resource "uptimekuma_notification_signl4" "example" {
  name        = "SIGNL4 Alerts"
  webhook_url = "https://connect.signl4.com/webhook/YOUR_TEAM_SECRET"
  is_active   = true
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_signl4" "write_only" {
  name = "SIGNL4 with a write-only secret"

  webhook_url_wo         = "https://connect.signl4.com/webhook/YOUR_TEAM_SECRET"
  webhook_url_wo_version = "1"
}
