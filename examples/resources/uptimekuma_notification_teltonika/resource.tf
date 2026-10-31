# Teltonika notification using the router's default modem
resource "uptimekuma_notification_teltonika" "sms" {
  name         = "Teltonika SMS"
  url          = "https://192.168.1.1"
  username     = "admin"
  password     = "your-router-password"
  phone_number = "+1234567890"
  is_default   = false
}

# Teltonika notification with a specific modem and a self-signed certificate
resource "uptimekuma_notification_teltonika" "self_signed" {
  name         = "Teltonika SMS (self-signed)"
  url          = "https://teltonika.example.com:8080"
  username     = "admin"
  password     = "your-router-password"
  modem        = "1-1"
  phone_number = "+1234567890,+1234567891"
  unsafe_tls   = true
}

# Requires Terraform or OpenTofu 1.11 or later: password_wo is never stored in
# plan or state. Change password_wo_version whenever the secret changes.
resource "uptimekuma_notification_teltonika" "write_only" {
  name = "Teltonika with a write-only secret"

  url          = "https://192.168.1.1"
  username     = "admin"
  phone_number = "+1234567890"

  password_wo         = "your-router-password"
  password_wo_version = "1"
}
