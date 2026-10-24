resource "uptimekuma_notification_flowtriq" "example" {
  name        = "Flowtriq Notifications"
  webhook_url = "https://app.flowtriq.com/api/webhooks/0123456789abcdef"
  is_active   = true
  is_default  = false
}

# Authenticate the webhook request. Uptime Kuma sends api_key as the "X-API-Key" header. If api_key
# is unset, the header is omitted.
resource "uptimekuma_notification_flowtriq" "authenticated" {
  name        = "Flowtriq Authenticated"
  webhook_url = "https://app.flowtriq.com/api/webhooks/0123456789abcdef"
  api_key     = "0123456789abcdef0123456789abcdef"
  is_active   = true
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo and api_key_wo are
# never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_flowtriq" "write_only" {
  name = "Flowtriq with write-only secrets"

  webhook_url_wo         = "https://app.flowtriq.com/api/webhooks/0123456789abcdef"
  webhook_url_wo_version = "1"
  api_key_wo             = "your-flowtriq-api-key"
  api_key_wo_version     = "1"
}
