resource "uptimekuma_notification_onebot" "group_message" {
  name         = "OneBot Group Notification"
  http_addr    = "http://localhost:5700"
  access_token = "your-access-token"
  msg_type     = "group"
  receiver_id  = "123456789"
  is_default   = false
}

resource "uptimekuma_notification_onebot" "private_message" {
  name         = "OneBot Private Notification"
  http_addr    = "http://localhost:5700"
  access_token = "your-access-token"
  msg_type     = "private"
  receiver_id  = "987654321"
  is_default   = false
}

# Requires Terraform or OpenTofu 1.11 or later: access_token_wo is never stored in
# plan or state. Change the matching *_wo_version whenever a secret changes.
resource "uptimekuma_notification_onebot" "write_only" {
  name        = "OneBot with write-only secret"
  http_addr   = "http://localhost:5700"
  msg_type    = "group"
  receiver_id = "123456789"

  access_token_wo         = "your-access-token"
  access_token_wo_version = "1"
}
