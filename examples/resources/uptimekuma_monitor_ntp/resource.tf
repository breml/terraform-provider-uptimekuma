# NTP Monitor Resource Example

# Minimal NTP monitor. Leaving port and the thresholds unset stores them as SQL
# NULL, which is what tells the check to apply its own fallbacks (port 123,
# stratum 5, offset 1000 ms, root dispersion 500 ms). Setting them to those same
# numbers is a different state, so prefer leaving them out.
resource "uptimekuma_monitor_ntp" "basic" {
  name     = "NTP Pool"
  hostname = "pool.ntp.org"

  # Public NTP servers rate-limit frequent queries, so the Uptime Kuma web UI
  # uses 300 seconds for NTP monitors. This provider keeps the shared default
  # of 60 seconds, so set it explicitly for public pool servers.
  interval = 300
}

# NTP monitor with all available options
resource "uptimekuma_monitor_ntp" "full" {
  name        = "NTP Cloudflare"
  description = "Stratum 3 time source for the build fleet"
  hostname    = "time.cloudflare.com"
  port        = 1123

  # The monitor goes down when the reported value reaches the threshold, so 4
  # already rejects stratum 4. 0 is rejected: the check reads it as unset.
  ntp_stratum_threshold         = 4
  ntp_time_offset_threshold     = 750
  ntp_root_dispersion_threshold = 250

  interval        = 300
  retry_interval  = 120
  resend_interval = 10
  max_retries     = 5
  active          = true
  upside_down     = false

  notification_ids = [1, 2]

  tags = [
    {
      tag_id = 1
      value  = "production"
    }
  ]
}

# NTP monitor with a fractional timeout, which round-trips unchanged
resource "uptimekuma_monitor_ntp" "fractional_timeout" {
  name     = "NTP Local Appliance"
  hostname = "10.0.0.1"
  timeout  = 2.5
  interval = 60
}
