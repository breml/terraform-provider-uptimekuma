# Example: Read PM2 monitor by name
data "uptimekuma_monitor_pm2" "example" {
  name = "API Server"
}

# Example: Read PM2 monitor by ID
data "uptimekuma_monitor_pm2" "by_id" {
  id = 1
}
