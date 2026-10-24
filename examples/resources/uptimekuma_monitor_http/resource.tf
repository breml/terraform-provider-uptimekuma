resource "uptimekuma_monitor_http" "example" {
  name                       = "Example API Monitoring"
  url                        = "https://api.example.com/health"
  interval                   = 60
  timeout                    = 30
  max_retries                = 2
  retry_interval             = 60
  resend_interval            = 0
  upside_down                = false
  active                     = true
  method                     = "GET"
  body                       = ""
  headers                    = ""
  auth_method                = ""
  basic_auth_user            = ""
  basic_auth_pass            = ""
  auth_domain                = ""
  auth_workstation           = ""
  ignore_tls                 = false
  tls_cert                   = ""
  tls_key                    = ""
  tls_ca                     = ""
  max_redirects              = 10
  accepted_status_codes      = ["200-299"]
  http_body_encoding         = "utf8"
  oauth_auth_method          = ""
  oauth_token_url            = ""
  oauth_client_id            = ""
  oauth_client_secret        = ""
  oauth_username             = ""
  oauth_password             = ""
  oauth_scope                = ""
  oauth_token_url_body       = ""
  expiry_notification        = false
  domain_expiry_notification = false
  ip_family                  = "auto"
}

resource "uptimekuma_monitor_http" "with_keyword" {
  name     = "HTTP with Keyword Validation"
  url      = "https://example.com"
  interval = 300
  timeout  = 30
  active   = true
}

resource "uptimekuma_monitor_http" "with_json_query" {
  name     = "HTTP with JSON Query"
  url      = "https://api.example.com/status"
  interval = 60
  timeout  = 30
  active   = true
  method   = "GET"
}

# Requires Terraform or OpenTofu 1.11 or later: basic_auth_pass_wo is never stored in
# plan or state. Change basic_auth_pass_wo_version whenever the secret changes.
resource "uptimekuma_monitor_http" "write_only" {
  name = "HTTP monitor with a write-only basic auth password"

  url         = "https://api.example.com/health"
  auth_method = "basic"

  basic_auth_user            = "monitor"
  basic_auth_pass_wo         = "basic-auth-password"
  basic_auth_pass_wo_version = "1"
}
