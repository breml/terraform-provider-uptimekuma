# Teltonika notification using the router's default modem
resource "uptimekuma_notification_teltonika" "sms" {
  name         = "Teltonika SMS"
  url          = "https://192.168.1.1"
  username     = "admin"
  password     = "your-router-password"
  phone_number = "+1234567890"
  is_active    = true
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
  is_active    = true
}
