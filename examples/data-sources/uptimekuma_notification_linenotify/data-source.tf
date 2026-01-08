# Data source to retrieve a LINE Notify notification by name
data "uptimekuma_notification_linenotify" "example_by_name" {
  name = "LINE Notify Alert"
}

# Data source to retrieve a LINE Notify notification by ID
data "uptimekuma_notification_linenotify" "example_by_id" {
  id = 1
}

output "notification_id" {
  value = data.uptimekuma_notification_linenotify.example_by_name.id
}

output "notification_name" {
  value = data.uptimekuma_notification_linenotify.example_by_id.name
}
