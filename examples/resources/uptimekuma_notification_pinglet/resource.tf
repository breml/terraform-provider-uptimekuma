resource "uptimekuma_notification_pinglet" "example" {
  name        = "Pinglet Notifications"
  publish_url = "https://app.pinglet.co.uk/your-namespace/alerts"
  api_key     = "your-pinglet-api-key"
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the API key changes.
resource "uptimekuma_notification_pinglet" "write_only" {
  name = "Pinglet with a write-only API key"

  publish_url = "https://app.pinglet.co.uk/your-namespace/alerts"

  api_key_wo         = "your-pinglet-api-key"
  api_key_wo_version = "1"
}
