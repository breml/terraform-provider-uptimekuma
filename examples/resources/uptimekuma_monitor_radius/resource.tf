# Radius Monitor Resource Example

# Basic Radius monitor with minimum required configuration
resource "uptimekuma_monitor_radius" "basic" {
  name            = "Radius Authentication Monitor"
  hostname        = "radius.example.com"
  radius_username = "monitor-user"
  radius_password = "monitor-password"
  radius_secret   = "shared-secret"
}

# Radius monitor with all available options
resource "uptimekuma_monitor_radius" "full" {
  name               = "Radius Authentication Monitor Full"
  description        = "Monitor Radius authentication server availability"
  hostname           = "radius.example.com"
  port               = 1812
  radius_username    = "monitor-user"
  radius_password    = "monitor-password"
  radius_secret      = "shared-secret"
  called_station_id  = "00-11-22-33-44-55"
  calling_station_id = "AA-BB-CC-DD-EE-FF"
  interval           = 120
  retry_interval     = 60
  max_retries        = 5
  active             = true
  upside_down        = false

  notification_ids = [1, 2]

  tags = [
    {
      tag_id = 1
      value  = "production"
    }
  ]
}

# Radius monitor on a non-default port (RADIUS accounting)
resource "uptimekuma_monitor_radius" "accounting" {
  name            = "Radius Accounting Monitor"
  hostname        = "radius.example.com"
  port            = 1813
  radius_username = "monitor-user"
  radius_password = "monitor-password"
  radius_secret   = "shared-secret"
  interval        = 300
}
