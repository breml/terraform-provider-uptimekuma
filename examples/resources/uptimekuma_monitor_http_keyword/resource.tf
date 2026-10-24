resource "uptimekuma_monitor_http_keyword" "example" {
  name                       = "Status Page Keyword Check"
  url                        = "https://status.example.com"
  keyword                    = "All systems operational"
  interval                   = 300
  timeout                    = 30
  method                     = "GET"
  max_retries                = 2
  upside_down                = false
  active                     = true
  ignore_tls                 = false
  max_redirects              = 10
  expiry_notification        = false
  domain_expiry_notification = false
  ip_family                  = "ipv4"
}

# Requires Terraform or OpenTofu 1.11 or later: basic_auth_pass_wo is never stored in
# plan or state. Change basic_auth_pass_wo_version whenever the secret changes.
resource "uptimekuma_monitor_http_keyword" "write_only" {
  name = "Keyword monitor with a write-only basic auth password"

  url         = "https://www.example.com"
  keyword     = "Welcome"
  auth_method = "basic"

  basic_auth_user            = "monitor"
  basic_auth_pass_wo         = "basic-auth-password"
  basic_auth_pass_wo_version = "1"
}
