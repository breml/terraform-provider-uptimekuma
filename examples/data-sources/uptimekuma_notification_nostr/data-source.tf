# Get Nostr notification by ID
data "uptimekuma_notification_nostr" "example_by_id" {
  id = 1
}

# Get Nostr notification by name
data "uptimekuma_notification_nostr" "example_by_name" {
  name = "Nostr Notification"
}

# Use the data source in a monitor
resource "uptimekuma_monitor_http" "example" {
  name = "Website Health Check"
  url  = "https://example.com"

  # Reference the Nostr notification by ID
  notification_ids = [data.uptimekuma_notification_nostr.example_by_id.id]
}
