# Read the current Uptime Kuma server settings
data "uptimekuma_settings" "current" {}

output "server_timezone" {
  value = data.uptimekuma_settings.current.server_timezone
}

output "keep_data_period_days" {
  value = data.uptimekuma_settings.current.keep_data_period_days
}
