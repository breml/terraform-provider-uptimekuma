resource "uptimekuma_monitor_ping" "example" {
  name           = "Server ICMP Monitoring"
  hostname       = "192.168.1.100"
  interval       = 60
  timeout        = 30
  max_retries    = 2
  retry_interval = 60
  upside_down    = false
  active         = true
  packet_size    = 56
}
