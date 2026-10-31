resource "uptimekuma_notification_splunk" "example" {
  name            = "Splunk Notification"
  rest_url        = "https://api.victorops.com/api/v2"
  severity        = "critical"
  auto_resolve    = "resolve"
  integration_key = "your_integration_key"
}

# Requires Terraform or OpenTofu 1.11 or later: rest_url_wo and integration_key_wo
# are never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_splunk" "write_only" {
  name = "Splunk with write-only secrets"

  severity     = "critical"
  auto_resolve = "resolve"

  rest_url_wo         = "https://api.victorops.com/api/v2"
  rest_url_wo_version = "1"

  integration_key_wo         = "your_integration_key"
  integration_key_wo_version = "1"
}
