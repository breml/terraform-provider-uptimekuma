resource "uptimekuma_notification_onebot" "group_message" {
  name         = "OneBot Group Notification"
  http_addr    = "http://localhost:5700"
  access_token = "your-access-token"
  msg_type     = "group"
  receiver_id  = "123456789"
  is_active    = true
  is_default   = false
}

resource "uptimekuma_notification_onebot" "private_message" {
  name         = "OneBot Private Notification"
  http_addr    = "http://localhost:5700"
  access_token = "your-access-token"
  msg_type     = "private"
  receiver_id  = "987654321"
  is_active    = true
  is_default   = false
}
