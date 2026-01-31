data "uptimekuma_notification_stackfield" "example" {
  name = "Stackfield Notifications"
}

output "stackfield_notification_id" {
  value = data.uptimekuma_notification_stackfield.example.id
}
