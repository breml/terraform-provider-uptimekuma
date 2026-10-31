resource "uptimekuma_notification_heiioncall" "example" {
  name       = "Heii On-Call Alerts"
  api_key    = "your-heii-oncall-api-key"
  trigger_id = "your-heii-oncall-trigger-id"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_heiioncall" "write_only" {
  name       = "Heii On-Call with write-only API key"
  trigger_id = "your-heii-oncall-trigger-id"

  api_key_wo         = "your-heii-oncall-api-key"
  api_key_wo_version = "1"
}
