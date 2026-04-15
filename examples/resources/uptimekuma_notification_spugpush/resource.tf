resource "uptimekuma_notification_spugpush" "example" {
  name         = "SpugPush Notifications"
  template_key = "your-template-key"
  is_active    = true
  is_default   = false
}
