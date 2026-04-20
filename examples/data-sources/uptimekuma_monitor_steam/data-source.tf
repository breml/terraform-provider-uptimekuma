# Look up a Steam game server monitor by name
data "uptimekuma_monitor_steam" "game_server" {
  name = "Game Server"
}

# Look up a Steam game server monitor by ID
data "uptimekuma_monitor_steam" "by_id" {
  id = 42
}

# Use the data source to reference an existing monitor
output "game_server_hostname" {
  value = data.uptimekuma_monitor_steam.game_server.hostname
}
