# Get a ClickSend SMS notification by name
data "uptimekuma_notification_clicksendsms" "example_by_name" {
  name = "ClickSend SMS Notification"
}

# Get a ClickSend SMS notification by ID
data "uptimekuma_notification_clicksendsms" "example_by_id" {
  id = 1
}

output "notification_name" {
  value = data.uptimekuma_notification_clicksendsms.example_by_name.name
}
