resource "uptimekuma_monitor_steam" "example" {
  name     = "Game Server Monitoring"
  hostname = "192.168.1.100"
  port     = 27015
  timeout  = 48
  interval = 60
  active   = true
}
