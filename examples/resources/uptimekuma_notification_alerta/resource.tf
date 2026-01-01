resource "uptimekuma_notification_alerta" "example" {
  name          = "Alerta Notifications"
  api_endpoint  = "https://alerta.example.com"
  api_key       = "your-alerta-api-key"
  environment   = "production"
  alert_state   = "alert"
  recover_state = "ok"
  is_active     = true
  is_default    = false
}

resource "uptimekuma_notification_alerta" "staging" {
  name          = "Alerta Staging"
  api_endpoint  = "https://alerta-staging.example.com"
  api_key       = "your-staging-api-key"
  environment   = "staging"
  alert_state   = "critical"
  recover_state = "resolved"
  is_active     = true
  is_default    = false
}
