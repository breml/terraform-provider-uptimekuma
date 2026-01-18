# Get PushDeer notification by ID
data "uptimekuma_notification_pushdeer" "by_id" {
  id = 1
}

# Get PushDeer notification by name
data "uptimekuma_notification_pushdeer" "by_name" {
  name = "PushDeer Notification"
}
