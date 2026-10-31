resource "uptimekuma_notification_whatsapp360messenger" "example" {
  name       = "WhatsApp 360messenger Notifications"
  auth_token = "YOUR_360MESSENGER_AUTH_TOKEN"
  recipient  = "+15551234567,+15557654321"
  is_default = false
}

resource "uptimekuma_notification_whatsapp360messenger" "with_template" {
  name         = "WhatsApp 360messenger with Template"
  auth_token   = "YOUR_360MESSENGER_AUTH_TOKEN"
  recipient    = "+15551234567"
  group_ids    = ["120363012345678901", "120363098765432109"]
  use_template = true
  template     = "{{ name }} is {{ status }}"
  is_default   = false
}

# Requires Terraform or OpenTofu 1.11 or later: auth_token_wo is never stored in
# plan or state. Change auth_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_whatsapp360messenger" "write_only" {
  name = "WhatsApp 360messenger with a write-only secret"

  recipient = "+15551234567"

  auth_token_wo         = "YOUR_360MESSENGER_AUTH_TOKEN"
  auth_token_wo_version = "1"
}
