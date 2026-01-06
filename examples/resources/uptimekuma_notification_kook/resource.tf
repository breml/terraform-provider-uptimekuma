resource "uptimekuma_notification_kook" "example" {
  name       = "Kook Notifications"
  bot_token  = "1/MzAxMjk5NzA1OTMzMDAwMA=="
  guild_id   = "382941547624206336"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_kook" "alerts" {
  name       = "Kook Alerts"
  bot_token  = "1/MzAxMjk5NzA1OTMzMDAwMA=="
  guild_id   = "382941547624206337"
  is_active  = true
  is_default = false
}
