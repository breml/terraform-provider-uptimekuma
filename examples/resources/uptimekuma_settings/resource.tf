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

# Requires Terraform or OpenTofu 1.11 or later: steam_api_key_wo is never stored in
# plan or state. Change steam_api_key_wo_version whenever the secret changes.
resource "uptimekuma_settings" "write_only" {
  steam_api_key_wo         = "steam-api-key"
  steam_api_key_wo_version = "1"
}
