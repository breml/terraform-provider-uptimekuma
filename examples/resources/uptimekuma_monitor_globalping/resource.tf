resource "uptimekuma_monitor_globalping" "example" {
  name      = "Globalping Ping Monitor"
  subtype   = "ping"
  url       = "https://example.com"
  location  = "Europe"
  ip_family = "ipv4"
  interval  = 60
  active    = true
}
