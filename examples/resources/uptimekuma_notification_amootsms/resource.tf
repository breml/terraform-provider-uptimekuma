# Send a plain SMS. line_number "public" is the value the Uptime Kuma web UI suggests.
resource "uptimekuma_notification_amootsms" "example" {
  name        = "Amoot SMS Notifications"
  api_token   = "0123456789abcdef0123456789abcdef"
  mobiles     = "09120000000,09120000001"
  line_number = "public"
  is_default  = false
}

# Send the message as the values of a pattern registered with Amoot SMS.
resource "uptimekuma_notification_amootsms" "pattern" {
  name            = "Amoot SMS Pattern"
  api_token       = "0123456789abcdef0123456789abcdef"
  mobiles         = "09120000000"
  use_pattern     = true
  pattern_code_id = 1234
}

# Send pattern messages from your own line.
resource "uptimekuma_notification_amootsms" "pattern_own_line" {
  name            = "Amoot SMS Pattern Own Line"
  api_token       = "0123456789abcdef0123456789abcdef"
  mobiles         = "09120000000"
  use_pattern     = true
  pattern_code_id = 1234
  use_own_line    = true
  line_number     = "50001"
}
