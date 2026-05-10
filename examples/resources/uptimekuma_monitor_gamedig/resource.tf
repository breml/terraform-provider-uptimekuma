resource "uptimekuma_monitor_gamedig" "example" {
  name                    = "Minecraft Server Monitoring"
  hostname                = "192.168.1.100"
  port                    = 25565
  game                    = "minecraft"
  gamedig_given_port_only = true
  interval                = 60
  active                  = true
}
