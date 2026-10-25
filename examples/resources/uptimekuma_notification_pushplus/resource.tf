resource "uptimekuma_notification_pushplus" "example" {
  name      = "PushPlus Notification"
  send_key  = "your-pushplus-send-key"
  is_active = true
}

# Requires Terraform or OpenTofu 1.11 or later: send_key_wo is never stored in
# plan or state. Change send_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_pushplus" "write_only" {
  name = "PushPlus with a write-only secret"

  send_key_wo         = "your-pushplus-send-key"
  send_key_wo_version = "1"
}
