resource "uptimekuma_monitor_push" "example" {
  name        = "External Push Event Monitoring"
  interval    = 3600
  timeout     = 30
  max_retries = 0
  upside_down = false
  active      = true
}

# push_token is sensitive, so any output that references it must be sensitive too.
output "push_url" {
  value     = "https://uptime.example.com/api/push/${uptimekuma_monitor_push.example.push_token}?status=up&msg=OK&ping="
  sensitive = true
}
