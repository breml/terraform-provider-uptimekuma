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
