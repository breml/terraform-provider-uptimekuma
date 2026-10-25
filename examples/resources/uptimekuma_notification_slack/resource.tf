resource "uptimekuma_notification_slack" "example" {
  name        = "Slack Notifications"
  webhook_url = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  channel     = "#monitoring"
  icon_emoji  = ":bell:"
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_slack" "with_mentions" {
  name           = "Slack with Mentions"
  webhook_url    = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  username       = "Uptime Kuma"
  icon_emoji     = ":warning:"
  channel_notify = true
  is_active      = true
  is_default     = false
}

# Requires Terraform or OpenTofu 1.11 or later: webhook_url_wo is never stored in
# plan or state. Change webhook_url_wo_version whenever the secret changes.
resource "uptimekuma_notification_slack" "write_only" {
  name = "Slack with a write-only secret"

  channel = "#monitoring"

  webhook_url_wo         = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  webhook_url_wo_version = "1"
}
