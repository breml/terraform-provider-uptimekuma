resource "uptimekuma_notification_ooredoo" "example" {
  name         = "Ooredoo Notifications"
  username     = "bulk-sms-user"
  access_key   = "0123456789abcdef"
  bearer_token = "0123456789abcdef0123456789abcdef"
  to_number    = "9601234567"
  is_active    = true
  is_default   = false
}

# Send to several recipients. Recipients are separated by comma, semicolon or whitespace, so an
# individual number must not contain spaces. Uptime Kuma strips every "+" character and prefixes
# bare 7 digit numbers with the Maldives country code "960".
resource "uptimekuma_notification_ooredoo" "multiple_recipients" {
  name         = "Ooredoo On-Call"
  username     = "bulk-sms-user"
  access_key   = "0123456789abcdef"
  bearer_token = "0123456789abcdef0123456789abcdef"
  to_number    = "1234567,+9607654321"
  is_active    = true
  is_default   = false
}

# Use a custom API endpoint. If server_url is unset, Uptime Kuma falls back to
# "https://o-papi1-lb01.ooredoo.mv/bulk_sms/v2".
resource "uptimekuma_notification_ooredoo" "custom_endpoint" {
  name         = "Ooredoo Custom Endpoint"
  username     = "bulk-sms-user"
  access_key   = "0123456789abcdef"
  bearer_token = "0123456789abcdef0123456789abcdef"
  to_number    = "9601234567"
  server_url   = "https://o-papi2-lb01.ooredoo.mv/bulk_sms/v2"
  is_active    = true
  is_default   = false
}

# Requires Terraform or OpenTofu 1.11 or later: access_key_wo and bearer_token_wo
# are never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_ooredoo" "write_only" {
  name      = "Ooredoo with write-only secrets"
  username  = "bulk-sms-user"
  to_number = "9601234567"

  access_key_wo           = "0123456789abcdef"
  access_key_wo_version   = "1"
  bearer_token_wo         = "0123456789abcdef0123456789abcdef"
  bearer_token_wo_version = "1"
}
