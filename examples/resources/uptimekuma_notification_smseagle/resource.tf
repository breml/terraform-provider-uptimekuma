# SMSEagle notification using API v2 with phone number recipient
resource "uptimekuma_notification_smseagle" "sms" {
  name           = "SMSEagle SMS"
  url            = "https://192.168.1.100"
  token          = "your-api-token"
  recipient_type = "smseagle-to"
  recipient_to   = "+1234567890"
  msg_type       = "smseagle-sms"
  api_type       = "smseagle-apiv2"
  is_active      = true
  is_default     = false
}

# SMSEagle notification with voice call (ring)
resource "uptimekuma_notification_smseagle" "ring" {
  name           = "SMSEagle Ring"
  url            = "https://smseagle.example.com"
  token          = "your-api-token"
  recipient_type = "smseagle-to"
  recipient_to   = "+1234567890"
  msg_type       = "smseagle-ring"
  duration       = 15
  api_type       = "smseagle-apiv2"
  is_active      = true
}
