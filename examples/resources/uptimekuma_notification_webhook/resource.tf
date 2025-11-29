resource "uptimekuma_notification_webhook" "example" {
  name                 = "Generic Webhook Notification"
  webhook_url          = "https://webhooks.example.com/uptime-kuma"
  webhook_content_type = "json"
  is_active            = true
  is_default           = false
}

resource "uptimekuma_notification_webhook" "with_headers" {
  name                 = "Webhook with Auth"
  webhook_url          = "https://webhooks.example.com/uptime-kuma"
  webhook_content_type = "json"
  webhook_additional_headers = {
    "Authorization"   = "Bearer token123"
    "X-Custom-Header" = "value"
  }
  is_active  = true
  is_default = false
}
