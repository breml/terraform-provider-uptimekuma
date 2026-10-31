# Post to a QQ group. Do not end http_addr with a slash: Uptime Kuma then skips
# the /api/ segment and posts to an endpoint that does not exist.
resource "uptimekuma_notification_milky" "group" {
  name         = "Milky Group Notifications"
  http_addr    = "http://localhost:3000"
  access_token = "your-milky-access-token"
  msg_type     = "group"
  receiver_id  = "123456789"
  is_active    = true
  is_default   = false
}

# Without msg_type, Uptime Kuma sends a private message to the user.
resource "uptimekuma_notification_milky" "private" {
  name         = "Milky Private Notifications"
  http_addr    = "http://localhost:3000"
  access_token = "your-milky-access-token"
  receiver_id  = "987654321"
}

# Requires Terraform or OpenTofu 1.11 or later: access_token_wo is never stored in
# plan or state. Change access_token_wo_version whenever the access token changes.
resource "uptimekuma_notification_milky" "write_only" {
  name = "Milky with a write-only access token"

  http_addr   = "http://localhost:3000"
  msg_type    = "group"
  receiver_id = "123456789"

  access_token_wo         = "your-milky-access-token"
  access_token_wo_version = "1"
}
