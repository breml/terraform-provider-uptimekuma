resource "uptimekuma_monitor_group" "example" {
  name   = "Production Services"
  status = "up"
  active = true
}

resource "uptimekuma_monitor_group" "nested" {
  name   = "Production - Web Services"
  parent = uptimekuma_monitor_group.example.id
  status = "up"
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
