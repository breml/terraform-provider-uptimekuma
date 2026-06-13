resource "uptimekuma_notification_whatsapp360messenger" "example" {
  name       = "WhatsApp 360messenger Notifications"
  auth_token = "YOUR_360MESSENGER_AUTH_TOKEN"
  recipient  = "+15551234567,+15557654321"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_whatsapp360messenger" "with_template" {
  name         = "WhatsApp 360messenger with Template"
  auth_token   = "YOUR_360MESSENGER_AUTH_TOKEN"
  recipient    = "+15551234567"
  group_ids    = ["120363012345678901", "120363098765432109"]
  use_template = true
  template     = "{{ name }} is {{ status }}"
  is_active    = true
  is_default   = false
}
