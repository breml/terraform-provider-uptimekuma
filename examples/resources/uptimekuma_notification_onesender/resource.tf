resource "uptimekuma_notification_onesender" "private" {
  name          = "OneSender Private"
  url           = "https://onesender.example.com/api/v1/send"
  token         = "your-api-token"
  receiver      = "+6281234567890"
  type_receiver = "private"
  is_active     = true
  is_default    = false
}

resource "uptimekuma_notification_onesender" "group" {
  name          = "OneSender Group"
  url           = "https://onesender.example.com/api/v1/send"
  token         = "your-api-token"
  receiver      = "group-id-123"
  type_receiver = "group"
  is_active     = true
  is_default    = false
}
