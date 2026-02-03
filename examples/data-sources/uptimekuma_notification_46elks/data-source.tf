data "uptimekuma_notification_46elks" "example" {
  name = "46elks SMS Notification"
}

output "notification_id" {
  value       = data.uptimekuma_notification_46elks.example.id
  description = "The ID of the 46elks notification"
}
