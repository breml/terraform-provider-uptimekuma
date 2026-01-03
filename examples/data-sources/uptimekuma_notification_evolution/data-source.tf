# Get an Evolution API notification by name
data "uptimekuma_notification_evolution" "example_by_name" {
  name = "Evolution WhatsApp Notification"
}

# Get an Evolution API notification by ID
data "uptimekuma_notification_evolution" "example_by_id" {
  id = 1
}

output "notification_name" {
  value = data.uptimekuma_notification_evolution.example_by_name.name
}
