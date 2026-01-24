# Get Pushy notification by ID
data "uptimekuma_notification_pushy" "by_id" {
  id = 1
}

# Get Pushy notification by name
data "uptimekuma_notification_pushy" "by_name" {
  name = "My Pushy Notification"
}

output "notification_name" {
  value = data.uptimekuma_notification_pushy.by_name.name
}
