# WAHA (WhatsApp HTTP API) notification resource example
# WAHA is an API for WhatsApp Business messaging
# Used for sending WhatsApp messages for monitoring alerts

resource "uptimekuma_notification_waha" "example" {
  name      = "WAHA Group Chat"
  is_active = true
  api_url   = "https://api.waha.local:3000"
  session   = "default"
  chat_id   = "120363101234567890@g.us"
  api_key   = "your-api-key-here" # Optional
}

resource "uptimekuma_notification_waha" "example_without_key" {
  name      = "WAHA Direct Message"
  is_active = true
  api_url   = "https://api.waha.local:3000"
  session   = "production"
  chat_id   = "5219123456789" # Phone number format for direct messages
}
