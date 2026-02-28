# Manage Uptime Kuma server settings
resource "uptimekuma_settings" "main" {
  server_timezone        = "Europe/Berlin"
  keep_data_period_days  = 30
  check_update           = true
  search_engine_index    = false
  entry_page             = "dashboard"
  nscd                   = true
  tls_expiry_notify_days = [7, 14, 21]
  trust_proxy            = false
}

# Minimal example: only manage specific settings
resource "uptimekuma_settings" "minimal" {
  server_timezone       = "UTC"
  keep_data_period_days = 7
}
