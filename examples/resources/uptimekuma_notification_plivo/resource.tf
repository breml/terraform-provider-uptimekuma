resource "uptimekuma_notification_plivo" "example" {
  name        = "Plivo Notifications"
  auth_id     = "MA0123456789ABCDEFGH"
  auth_token  = "0123456789abcdef0123456789abcdef"
  from_number = "+15550001111"
  to_number   = "+15550002222"
  is_active   = true
  is_default  = false
}

# Deliver the alert as a voice call. Plivo fetches answer_url with an HTTP GET to obtain the
# Plivo XML driving the call, with the alert text set as the "message" query parameter,
# replacing any "message" parameter already present.
resource "uptimekuma_notification_plivo" "voice_call" {
  name         = "Plivo Voice Call"
  auth_id      = "MA0123456789ABCDEFGH"
  auth_token   = "0123456789abcdef0123456789abcdef"
  from_number  = "+15550001111"
  to_number    = "+15550002222"
  message_type = "call"
  answer_url   = "https://example.com/plivo/answer"
  is_active    = true
  is_default   = false
}
