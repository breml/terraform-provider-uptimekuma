data "uptimekuma_notification_smtp" "by_id" {
  id = 1
}

data "uptimekuma_notification_smtp" "by_name" {
  name = "SMTP Notifications"
}
