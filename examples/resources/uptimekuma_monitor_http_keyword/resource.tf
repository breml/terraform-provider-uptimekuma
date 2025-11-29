resource "uptimekuma_monitor_http_keyword" "example" {
  name                = "Status Page Keyword Check"
  url                 = "https://status.example.com"
  keyword             = "All systems operational"
  interval            = 300
  timeout             = 30
  method              = "GET"
  max_retries         = 2
  upside_down         = false
  active              = true
  ignore_tls          = false
  max_redirects       = 10
  expiry_notification = true
}
