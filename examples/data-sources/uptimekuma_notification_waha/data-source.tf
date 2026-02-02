# WAHA notification data source examples
# Get notification information by ID or name

# Get WAHA notification by ID
data "uptimekuma_notification_waha" "by_id" {
  id = 1
}

# Get WAHA notification by name
data "uptimekuma_notification_waha" "by_name" {
  name = "WAHA Group Chat"
}

# Use the data source output
output "waha_notification_name" {
  value = data.uptimekuma_notification_waha.by_name.name
}
