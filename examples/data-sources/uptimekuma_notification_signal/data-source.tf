data "uptimekuma_notification_signal" "example" {
  name = "My Signal Notification"
}

output "notification_id" {
  value = data.uptimekuma_notification_signal.example.id
}
