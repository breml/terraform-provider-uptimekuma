# Send the default message to a list of WhatsApp chats.
resource "uptimekuma_notification_openwa" "example" {
  name       = "OpenWA Notifications"
  api_url    = "http://localhost:2785/"
  api_key    = "your-openwa-api-key"
  session    = "default"
  chat_id    = "00117612345678@c.us,123456789012345678@g.us,1234567890@lid"
  is_default = false
}

# Send a custom message rendered from a template.
resource "uptimekuma_notification_openwa" "custom_message" {
  name               = "OpenWA Custom Message"
  api_url            = "http://localhost:2785/"
  api_key            = "your-openwa-api-key"
  session            = "default"
  chat_id            = "00117612345678@c.us"
  use_custom_message = true
  custom_message     = "{{ name }} is {{ status }}: {{ msg }}"
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo is never stored in
# plan or state. Change api_key_wo_version whenever the API key changes.
resource "uptimekuma_notification_openwa" "write_only" {
  name = "OpenWA with a write-only API key"

  api_url = "http://localhost:2785/"
  session = "default"
  chat_id = "00117612345678@c.us"

  api_key_wo         = "your-openwa-api-key"
  api_key_wo_version = "1"
}
