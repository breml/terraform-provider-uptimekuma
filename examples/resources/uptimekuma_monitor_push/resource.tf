resource "uptimekuma_monitor_push" "example" {
  name        = "External Push Event Monitoring"
  push_token  = "auto_generated_by_uptime_kuma"
  interval    = 3600
  timeout     = 30
  max_retries = 0
  upside_down = false
  active      = true
}
