resource "uptimekuma_monitor_tcp_port" "example" {
  name        = "Database TCP Port Monitoring"
  hostname    = "db.example.com"
  port        = 5432
  interval    = 60
  timeout     = 30
  max_retries = 2
  upside_down = false
  active      = true
}
