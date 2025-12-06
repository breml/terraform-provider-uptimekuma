# Associate monitors with a maintenance window
resource "uptimekuma_maintenance" "weekly" {
  title       = "Weekly Maintenance"
  description = "Regular weekly maintenance window"
  strategy    = "recurring-weekday"
  active      = true

  weekdays = [1]

  start_time = {
    hours   = 2
    minutes = 0
    seconds = 0
  }

  end_time = {
    hours   = 4
    minutes = 0
    seconds = 0
  }

  timezone = "UTC"
}

resource "uptimekuma_monitor_http" "example" {
  name = "Example Website"
  url  = "https://example.com"
}

resource "uptimekuma_maintenance_monitors" "example" {
  maintenance_id = uptimekuma_maintenance.weekly.id
  monitor_ids    = [uptimekuma_monitor_http.example.id]
}
