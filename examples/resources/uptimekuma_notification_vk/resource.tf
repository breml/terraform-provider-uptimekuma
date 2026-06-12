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
