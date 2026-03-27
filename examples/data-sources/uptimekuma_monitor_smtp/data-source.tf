# Query SMTP monitor by ID.
data "uptimekuma_monitor_smtp" "example_by_id" {
  id = 42
}

# Query SMTP monitor by name.
# Returns the monitor ID and hostname for use in other resources.
data "uptimekuma_monitor_smtp" "example_by_name" {
  name = "Mail Server"
}
