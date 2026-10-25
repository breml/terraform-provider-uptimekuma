resource "uptimekuma_notification_spugpush" "example" {
  name         = "SpugPush Notifications"
  template_key = "your-template-key"
  is_active    = true
  is_default   = false
}

# Requires Terraform or OpenTofu 1.11 or later: template_key_wo is never stored in
# plan or state. Change template_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_spugpush" "write_only" {
  name = "SpugPush with a write-only secret"

  template_key_wo         = "your-template-key"
  template_key_wo_version = "1"
}
