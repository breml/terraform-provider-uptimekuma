resource "uptimekuma_notification_signal" "example" {
  name       = "Signal Notifications"
  url        = "http://signal.example.com:8080"
  number     = "+1234567890"
  recipients = "+9876543210,+1111111111"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_signal" "with_single_recipient" {
  name       = "Signal Single Recipient"
  url        = "http://signal.example.com:8080"
  number     = "+1234567890"
  recipients = "+9876543210"
  is_active  = true
  is_default = false
}
