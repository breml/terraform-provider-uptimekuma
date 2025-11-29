resource "uptimekuma_monitor_dns" "example" {
  name               = "DNS Resolution Check"
  hostname           = "example.com"
  dns_resolve_server = "8.8.8.8"
  interval           = 300
  timeout            = 30
  max_retries        = 2
  upside_down        = false
  active             = true
  port               = 53
}
