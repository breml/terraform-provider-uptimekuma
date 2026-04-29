resource "uptimekuma_monitor_tailscale_ping" "example" {
  name           = "Tailscale Node Monitoring"
  hostname       = "100.64.0.1"
  interval       = 60
  max_retries    = 2
  retry_interval = 60
  upside_down    = false
  active         = true
}
