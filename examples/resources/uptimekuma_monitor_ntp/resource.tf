resource "uptimekuma_monitor_ntp" "example" {
  name     = "NTP Pool"
  hostname = "pool.ntp.org"
  port     = 123
  timeout  = 10

  ntp_stratum_threshold         = 5
  ntp_time_offset_threshold     = 1000
  ntp_root_dispersion_threshold = 500

  # Public NTP servers rate-limit frequent queries, so the Uptime Kuma web UI
  # suggests 300 seconds for NTP monitors.
  interval = 300
  active   = true
}
