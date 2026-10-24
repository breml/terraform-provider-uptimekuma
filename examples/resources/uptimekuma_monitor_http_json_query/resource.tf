resource "uptimekuma_monitor_http_json_query" "example" {
  name                       = "API JSON Response Check"
  url                        = "https://api.example.com/status"
  json_path                  = "$.status"
  json_path_expected         = "ok"
  interval                   = 60
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

# Requires Terraform or OpenTofu 1.11 or later: oauth_client_secret_wo is never stored in
# plan or state. Change oauth_client_secret_wo_version whenever the secret changes.
resource "uptimekuma_monitor_http_json_query" "write_only" {
  name = "JSON query monitor with a write-only OAuth client secret"

  url            = "https://api.example.com/status"
  json_path      = "$.status"
  expected_value = "ok"
  auth_method    = "oauth2-cc"

  oauth_auth_method              = "client_secret_basic"
  oauth_token_url                = "https://auth.example.com/oauth/token"
  oauth_client_id                = "uptime-kuma"
  oauth_client_secret_wo         = "oauth-client-secret"
  oauth_client_secret_wo_version = "1"
}
