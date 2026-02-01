# Threema notification resource example
# Threema is a secure messaging application with end-to-end encryption
# Used for sending encrypted monitoring alerts

resource "uptimekuma_notification_threema" "example_email" {
  name            = "Threema Email Recipient"
  is_active       = true
  sender_identity = "YOUR_GATEWAY_ID"
  secret          = "YOUR_API_SECRET" # Sensitive field
  recipient       = "user@example.com"
  recipient_type  = "email"
}

resource "uptimekuma_notification_threema" "example_phone" {
  name            = "Threema Phone Recipient"
  is_active       = true
  sender_identity = "YOUR_GATEWAY_ID"
  secret          = "YOUR_API_SECRET" # Sensitive field
  recipient       = "+41791234567"
  recipient_type  = "phone"
}

resource "uptimekuma_notification_threema" "example_identity" {
  name            = "Threema Identity Recipient"
  is_active       = true
  sender_identity = "YOUR_GATEWAY_ID"
  secret          = "YOUR_API_SECRET" # Sensitive field
  recipient       = "ECHOECHO"
  recipient_type  = "identity"
}
