resource "uptimekuma_notification_fluxer" "example" {
  name        = "Fluxer Notifications"
  webhook_url = "https://fluxer.example.com/webhook/YOUR_TOKEN"
  username    = "Uptime Kuma"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_fluxer" "with_template" {
  name                 = "Fluxer with Template"
  webhook_url          = "https://fluxer.example.com/webhook/YOUR_TOKEN"
  prefix_message       = "Alert:"
  disable_url          = true
  use_message_template = true
  message_format       = "markdown"
  message_template     = "{{ name }} is {{ status }}"
  is_active            = true
  is_default           = false
}
