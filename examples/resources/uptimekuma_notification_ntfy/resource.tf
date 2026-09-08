resource "uptimekuma_notification_ntfy" "example" {
  name       = "ntfy.sh Notifications"
  topic      = "my_uptime_kuma_notifications"
  priority   = 5
  server_url = "https://ntfy.sh"
  icon       = "https://example.com/icon.png"
  is_active  = true
  is_default = false
}

resource "uptimekuma_notification_ntfy" "high_priority" {
  name       = "ntfy Critical Alerts"
  topic      = "critical_alerts"
  priority   = 1
  server_url = "https://ntfy.sh"
  is_active  = true
  is_default = true
}

# Self hosted ntfy server, authenticated with an access token.
resource "uptimekuma_notification_ntfy" "access_token" {
  name       = "ntfy Self Hosted"
  topic      = "uptime_kuma"
  server_url = "https://ntfy.example.com"

  authentication_method = "accessToken"
  access_token          = var.ntfy_access_token

  priority      = 5
  priority_down = 1
}

# Self hosted ntfy server, authenticated with username and password, using
# custom title and message templates.
resource "uptimekuma_notification_ntfy" "username_password" {
  name       = "ntfy Self Hosted (Basic Auth)"
  topic      = "uptime_kuma"
  server_url = "https://ntfy.example.com"

  authentication_method = "usernamePassword"
  username              = var.ntfy_username
  password              = var.ntfy_password

  use_template   = true
  custom_title   = "{{ name }} is {{ status }}"
  custom_message = "{{ msg }}"
}
