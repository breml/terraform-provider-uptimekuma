resource "uptimekuma_notification_matrix" "example" {
  name             = "Matrix Example"
  is_active        = true
  homeserver_url   = "https://matrix.example.com"
  internal_room_id = "!abc123:example.com"
  access_token     = "syt_access_token_example"
}
