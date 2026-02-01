# Threema notification data source examples
# Get notification information by ID or name

# Get Threema notification by ID
data "uptimekuma_notification_threema" "by_id" {
  id = 1
}

# Get Threema notification by name
data "uptimekuma_notification_threema" "by_name" {
  name = "Threema Email Recipient"
}

# Use the data source output
output "threema_notification_name" {
  value = data.uptimekuma_notification_threema.by_name.name
}
