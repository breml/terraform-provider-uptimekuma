resource "uptimekuma_monitor_real_browser" "example" {
  name        = "Web Application Monitor"
  description = "Monitor web application using real browser"
  url         = "https://example.com"
  interval    = 60
  timeout     = 48
  active      = true
}

resource "uptimekuma_monitor_real_browser" "example_with_custom_status_codes" {
  name                  = "Custom Status Code Monitor"
  description           = "Monitor with custom accepted status codes"
  url                   = "https://example.com/app"
  interval              = 120
  timeout               = 60
  accepted_status_codes = ["200-299", "301", "302"]
  ignore_tls            = false
  max_redirects         = 10
  active                = true
}

resource "uptimekuma_monitor_real_browser" "example_with_remote_browser" {
  name           = "Remote Browser Monitor"
  description    = "Monitor using a remote browser instance"
  url            = "https://example.com/dashboard"
  interval       = 180
  timeout        = 90
  remote_browser = 1
  active         = true
}
