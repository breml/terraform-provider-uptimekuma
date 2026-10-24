resource "uptimekuma_monitor_globalping" "example" {
  name      = "Globalping Ping Monitor"
  subtype   = "ping"
  url       = "https://example.com"
  location  = "Europe"
  ip_family = "ipv4"
  interval  = 60
  active    = true
}

# Requires Terraform or OpenTofu 1.11 or later: basic_auth_pass_wo is never stored in
# plan or state. Change basic_auth_pass_wo_version whenever the secret changes.
resource "uptimekuma_monitor_globalping" "write_only" {
  name = "Globalping monitor with a write-only basic auth password"

  subtype     = "http"
  url         = "https://www.example.com"
  auth_method = "basic"

  basic_auth_user            = "monitor"
  basic_auth_pass_wo         = "basic-auth-password"
  basic_auth_pass_wo_version = "1"
}
