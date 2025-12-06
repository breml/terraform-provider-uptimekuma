# Single maintenance window
resource "uptimekuma_maintenance" "example_single" {
  title       = "Server Upgrade"
  description = "Planned server upgrade for maintenance"
  strategy    = "single"
  active      = true

  start_date = "2024-12-01T02:00:00Z"
  end_date   = "2024-12-01T04:00:00Z"

  timezone = "UTC"
}

# Recurring weekly maintenance (every Monday and Wednesday)
resource "uptimekuma_maintenance" "example_weekly" {
  title       = "Weekly Maintenance"
  description = "Regular weekly maintenance window"
  strategy    = "recurring-weekday"
  active      = true

  weekdays = [1, 3]

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

  timezone = "America/New_York"
}

# Recurring interval (every 3 days)
resource "uptimekuma_maintenance" "example_interval" {
  title       = "Regular Backup Window"
  description = "Maintenance window for backups"
  strategy    = "recurring-interval"
  active      = true

  interval_day = 3

  start_time = {
    hours   = 1
    minutes = 0
    seconds = 0
  }

  end_time = {
    hours   = 3
    minutes = 0
    seconds = 0
  }

  timezone = "UTC"
}

# Monthly maintenance (1st and 15th of each month)
resource "uptimekuma_maintenance" "example_monthly" {
  title       = "Monthly Patches"
  description = "Monthly security patches"
  strategy    = "recurring-day-of-month"
  active      = true

  days_of_month = ["1", "15"]

  start_time = {
    hours   = 3
    minutes = 0
    seconds = 0
  }

  end_time = {
    hours   = 5
    minutes = 0
    seconds = 0
  }

  timezone = "UTC"
}

# Cron-based maintenance
resource "uptimekuma_maintenance" "example_cron" {
  title       = "Daily Early Morning Maintenance"
  description = "Daily maintenance at 2 AM"
  strategy    = "cron"
  active      = true

  cron             = "0 2 * * *"
  duration_minutes = 60

  timezone = "UTC"
}

# Manual maintenance
resource "uptimekuma_maintenance" "example_manual" {
  title       = "Emergency Maintenance"
  description = "On-demand emergency maintenance"
  strategy    = "manual"
  active      = false
}
