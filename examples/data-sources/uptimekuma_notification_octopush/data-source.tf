# Get Octopush notification by ID
data "uptimekuma_notification_octopush" "example_by_id" {
  id = 1
}

# Get Octopush notification by name
data "uptimekuma_notification_octopush" "example_by_name" {
  name = "Octopush SMS V2"
}

# Use the data source in a monitor
resource "uptimekuma_monitor_http" "example" {
  name = "Website Health Check"
  url  = "https://example.com"

  # Reference the Octopush notification by ID
  notification_ids = [data.uptimekuma_notification_octopush.example_by_id.id]
}
