# Look up a Google Chat notification by ID
data "uptimekuma_notification_googlechat" "example_by_id" {
  id = 1
}

# Look up a Google Chat notification by name
data "uptimekuma_notification_googlechat" "example_by_name" {
  name = "My Google Chat Notification"
}
