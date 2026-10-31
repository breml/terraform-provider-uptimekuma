resource "uptimekuma_notification_opsgenie" "example" {
  name     = "OpsGenie Example"
  api_key  = "your-opsgenie-api-key"
  region   = "us"
  priority = 2
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in plan
# or state. Change the matching *_wo_version whenever a secret changes.
resource "uptimekuma_notification_opsgenie" "write_only" {
  name   = "OpsGenie with write-only secret"
  region = "us"

  api_key_wo         = "your-opsgenie-api-key"
  api_key_wo_version = "1"
}
