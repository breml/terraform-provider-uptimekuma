# Look up a Globalping monitor by name
data "uptimekuma_monitor_globalping" "by_name" {
  name = "Globalping Ping Monitor"
}

# Look up a Globalping monitor by ID
data "uptimekuma_monitor_globalping" "by_id" {
  id = 42
}

# Use the data source to reference an existing monitor
output "globalping_subtype" {
  value = data.uptimekuma_monitor_globalping.by_name.subtype
}
