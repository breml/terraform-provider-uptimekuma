resource "uptimekuma_monitor_group" "example" {
  name   = "Production Services"
  active = true
}

resource "uptimekuma_monitor_group" "nested" {
  name   = "Production - Web Services"
  parent = uptimekuma_monitor_group.example.id
  active = true
}

resource "uptimekuma_monitor_http" "in_group" {
  name     = "API in Group"
  url      = "https://api.example.com/health"
  interval = 60
  timeout  = 30
  active   = true
  parent   = uptimekuma_monitor_group.nested.id
}
