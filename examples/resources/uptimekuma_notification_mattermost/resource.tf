resource "uptimekuma_notification_mattermost" "example" {
  name        = "Mattermost Example"
  webhook_url = "https://mattermost.example.com/hooks/xxx"
  username    = "Uptime Kuma"
  channel     = "#alerts"
  icon_emoji  = ":robot_face:"
  icon_url    = "https://example.com/icon.png"
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_mattermost" "write_only" {
  name = "Mattermost with write-only webhook URL"

  webhook_url_wo         = "https://mattermost.example.com/hooks/xxx"
  webhook_url_wo_version = "1"
}
