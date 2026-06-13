resource "uptimekuma_monitor_system_service" "example" {
  name                = "Nginx Service"
  system_service_name = "nginx.service"
  interval            = 60
  max_retries         = 2
  upside_down         = false
  active              = true
}
