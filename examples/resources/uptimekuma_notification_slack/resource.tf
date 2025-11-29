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
