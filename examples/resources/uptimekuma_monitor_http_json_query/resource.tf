resource "uptimekuma_monitor_http_json_query" "example" {
  name               = "API Status Check"
  description        = "Monitor API JSON response for status field"
  url                = "https://api.example.com/health"
  json_path          = "$.status"
  expected_value     = "ok"
  json_path_operator = "=="
  interval           = 60
  active             = true
}

resource "uptimekuma_monitor_http_json_query" "example_numeric" {
  name               = "API Response Time Check"
  description        = "Monitor API response time is under threshold"
  url                = "https://api.example.com/metrics"
  json_path          = "$.response_time_ms"
  expected_value     = "500"
  json_path_operator = "<"
  interval           = 120
  active             = true
}

resource "uptimekuma_monitor_http_json_query" "example_contains" {
  name               = "API Version Check"
  description        = "Monitor API version contains expected string"
  url                = "https://api.example.com/version"
  json_path          = "$.version"
  expected_value     = "v2"
  json_path_operator = "contains"
  interval           = 300
  active             = true
}
