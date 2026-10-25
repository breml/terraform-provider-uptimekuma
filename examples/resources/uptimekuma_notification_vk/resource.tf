resource "uptimekuma_notification_vk" "example" {
  name         = "VK Notifications"
  access_token = "vk1.a.YOUR_ACCESS_TOKEN"
  peer_id      = "12345"
  is_active    = true
  is_default   = false
}

resource "uptimekuma_notification_vk" "custom" {
  name             = "VK Notifications (custom API version)"
  access_token     = "vk1.a.YOUR_ACCESS_TOKEN"
  peer_id          = "2000000001"
  api_version      = "5.131"
  dont_parse_links = true
  is_active        = true
  is_default       = false
}

# Requires Terraform or OpenTofu 1.11 or later: access_token_wo is never stored in
# plan or state. Change access_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_vk" "write_only" {
  name = "VK with a write-only secret"

  peer_id = "12345"

  access_token_wo         = "vk1.a.YOUR_ACCESS_TOKEN"
  access_token_wo_version = "1"
}
