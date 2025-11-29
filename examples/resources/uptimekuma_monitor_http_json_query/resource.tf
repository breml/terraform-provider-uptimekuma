resource "uptimekuma_monitor_http_json_query" "example" {
  name                = "API JSON Response Check"
  url                 = "https://api.example.com/status"
  json_path           = "$.status"
  json_path_expected  = "ok"
  interval            = 60
  timeout             = 30
  method              = "GET"
  max_retries         = 2
  upside_down         = false
  active              = true
  ignore_tls          = false
  max_redirects       = 10
  expiry_notification = true
}
