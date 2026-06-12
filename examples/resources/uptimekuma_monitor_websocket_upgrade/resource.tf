resource "uptimekuma_monitor_websocket_upgrade" "example" {
  name                                  = "Websocket Upgrade Monitoring"
  url                                   = "wss://echo.websocket.org"
  ws_subprotocol                        = "chat"
  ws_ignore_sec_websocket_accept_header = false
  interval                              = 60
  max_retries                           = 2
  upside_down                           = false
  active                                = true
}
