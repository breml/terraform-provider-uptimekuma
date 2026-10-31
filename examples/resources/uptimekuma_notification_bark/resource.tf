resource "uptimekuma_notification_bark" "example" {
  name        = "Bark Notification"
  endpoint    = "https://api.bark.com"
  group       = "monitoring"
  sound       = "default"
  api_version = "v1"
}

# Requires Terraform or OpenTofu 1.11 or later: endpoint_wo is never stored in
# plan or state. Change endpoint_wo_version whenever the secret changes.
resource "uptimekuma_notification_bark" "write_only" {
  name = "Bark with a write-only secret"

  endpoint_wo         = "https://api.day.app/your-device-key"
  endpoint_wo_version = "1"
}
