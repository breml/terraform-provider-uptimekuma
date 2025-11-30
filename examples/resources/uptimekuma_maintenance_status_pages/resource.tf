# Associate status pages with a maintenance window
resource "uptimekuma_maintenance" "monthly" {
  title       = "Monthly Maintenance"
  description = "Monthly maintenance window"
  strategy    = "recurring-day-of-month"
  active      = true

  days_of_month = ["1"]

  start_time = {
    hours   = 0
    minutes = 0
    seconds = 0
  }

  end_time = {
    hours   = 2
    minutes = 0
    seconds = 0
  }

  timezone = "UTC"
}

resource "uptimekuma_maintenance_status_pages" "example" {
  maintenance_id  = uptimekuma_maintenance.monthly.id
  status_page_ids = [1, 2]
}
