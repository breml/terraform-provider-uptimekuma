# Check a PM2 process by name (recommended: PM2 reassigns numeric ids)
resource "uptimekuma_monitor_pm2" "example" {
  name         = "API Server"
  process_name = "api-server"
  interval     = 60
  max_retries  = 2
  active       = true
}

# The numeric PM2 id works as well, passed as a string
resource "uptimekuma_monitor_pm2" "by_id" {
  name         = "Worker 0"
  description  = "PM2 process with id 0"
  process_name = "0"
  interval     = 120
}
