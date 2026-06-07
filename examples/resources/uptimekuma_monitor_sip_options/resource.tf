resource "uptimekuma_monitor_sip_options" "example" {
  name        = "SIP Server Monitoring"
  hostname    = "sip.example.com"
  port        = 5060
  interval    = 60
  max_retries = 2
  upside_down = false
  active      = true
}
