resource "uptimekuma_notification_matrix" "example" {
  name             = "Matrix Example"
  homeserver_url   = "https://matrix.example.com"
  internal_room_id = "!abc123:example.com"
  access_token     = "syt_access_token_example"
}

# Requires Terraform or OpenTofu 1.11 or later: access_token_wo is never stored
# in plan or state. Change access_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_matrix" "write_only" {
  name             = "Matrix with write-only token"
  homeserver_url   = "https://matrix.example.com"
  internal_room_id = "!abc123:example.com"

  access_token_wo         = "syt_access_token_example"
  access_token_wo_version = "1"
}
