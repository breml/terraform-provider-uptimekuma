resource "uptimekuma_notification_telnyx" "example" {
  name         = "Telnyx Notifications"
  api_key      = "KEY0123456789ABCDEF"
  phone_number = "+15550001111"
  to_number    = "+15550002222"
  is_active    = true
  is_default   = false
}

resource "uptimekuma_notification_telnyx" "with_messaging_profile" {
  name                 = "Telnyx with Messaging Profile"
  api_key              = "KEY0123456789ABCDEF"
  messaging_profile_id = "40017a13-3f93-4d2d-b29e-1a000000000a"
  phone_number         = "+15550001111"
  to_number            = "+15550002222"
  is_active            = true
  is_default           = false
}
