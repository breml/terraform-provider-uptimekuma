resource "uptimekuma_monitor_websocket_upgrade" "example" {
  name                                  = "Websocket Upgrade Monitoring"
  url                                   = "wss://echo.websocket.org"
  ws_subprotocol                        = "chat"
  ws_ignore_sec_websocket_accept_header = false
  interval                              = 60
  max_retries                           = 2
  upside_down                           = false
  active                                = true
  domain_expiry_notification            = false
}

# Requires Terraform or OpenTofu 1.11 or later: basic_auth_pass_wo is never stored in
# plan or state. Change basic_auth_pass_wo_version whenever the secret changes.
resource "uptimekuma_monitor_websocket_upgrade" "write_only" {
  name = "WebSocket monitor with a write-only basic auth password"

  url         = "wss://ws.example.com/socket"
  auth_method = "basic"

  basic_auth_user            = "monitor"
  basic_auth_pass_wo         = "basic-auth-password"
  basic_auth_pass_wo_version = "1"
}
