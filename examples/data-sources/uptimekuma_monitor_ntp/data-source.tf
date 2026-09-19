# Look up an NTP monitor by name
data "uptimekuma_monitor_ntp" "pool" {
  name = "NTP Pool"
}

# Look up an NTP monitor by ID
data "uptimekuma_monitor_ntp" "by_id" {
  id = 42
}

# Use the data source to reference an existing monitor
output "ntp_hostname" {
  value = data.uptimekuma_monitor_ntp.pool.hostname
}
