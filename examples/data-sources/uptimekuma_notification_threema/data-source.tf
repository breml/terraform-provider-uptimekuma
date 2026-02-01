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

# Use with a resource
data "uptimekuma_notification_threema" "example" {
  name = uptimekuma_notification_threema.example_email.name
}
