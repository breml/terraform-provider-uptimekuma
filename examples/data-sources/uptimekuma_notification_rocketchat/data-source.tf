data "uptimekuma_notification_rocketchat" "example" {
  name = "My RocketChat Notification"
}

output "notification_id" {
  value = data.uptimekuma_notification_rocketchat.example.id
}
