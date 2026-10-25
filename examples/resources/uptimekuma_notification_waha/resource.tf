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

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_waha" "write_only" {
  name = "WAHA with a write-only secret"

  api_url = "https://api.waha.local:3000"
  session = "default"
  chat_id = "120363101234567890@g.us"

  api_key_wo         = "your-api-key-here"
  api_key_wo_version = "1"
}
