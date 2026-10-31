resource "uptimekuma_notification_fluxer" "example" {
  name        = "Fluxer Notifications"
  webhook_url = "https://fluxer.example.com/webhook/YOUR_TOKEN"
  username    = "Uptime Kuma"
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
  is_default           = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_fluxer" "write_only" {
  name = "Fluxer with a write-only secret"

  webhook_url_wo         = "https://fluxer.example.com/webhooks/your-token"
  webhook_url_wo_version = "1"
}
