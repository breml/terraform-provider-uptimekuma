# Look up a GameDig game server monitor by name
data "uptimekuma_monitor_gamedig" "minecraft_server" {
  name = "Minecraft Server"
}

# Look up a GameDig game server monitor by ID
data "uptimekuma_monitor_gamedig" "by_id" {
  id = 42
}

# Use the data source to reference an existing monitor
output "minecraft_server_hostname" {
  value = data.uptimekuma_monitor_gamedig.minecraft_server.hostname
}
