# List all maintenance windows
data "uptimekuma_maintenances" "all" {}

# Output the maintenance windows
output "maintenance_count" {
  value = length(data.uptimekuma_maintenances.all.maintenances)
}

output "maintenance_titles" {
  value = [for m in data.uptimekuma_maintenances.all.maintenances : m.title]
}
