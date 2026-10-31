resource "uptimekuma_monitor_tcp_port" "example" {
  name                       = "Database TCP Port Monitoring"
  hostname                   = "db.example.com"
  port                       = 5432
  interval                   = 60
  timeout                    = 30
  max_retries                = 2
  upside_down                = false
  active                     = true
  domain_expiry_notification = false
}

# TCP port monitor that performs a TLS handshake and checks the certificate.
# Uptime Kuma only runs the TLS check while expiry_notification is true.
resource "uptimekuma_monitor_tcp_port" "tls" {
  name                = "IMAPS TLS Monitoring"
  hostname            = "mail.example.com"
  port                = 993
  smtp_security       = "secure"
  expiry_notification = true
}
