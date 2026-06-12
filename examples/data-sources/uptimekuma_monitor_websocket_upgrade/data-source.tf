# Example: Read Websocket Upgrade monitor by name
data "uptimekuma_monitor_websocket_upgrade" "example" {
  name = "Websocket Upgrade Monitoring"
}

# Example: Read Websocket Upgrade monitor by ID
data "uptimekuma_monitor_websocket_upgrade" "by_id" {
  id = 1
}
