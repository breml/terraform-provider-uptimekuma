resource "uptimekuma_notification_evolution" "example" {
  name          = "Evolution WhatsApp Notification"
  is_active     = true
  api_url       = "https://api.evolution.example.com"
  instance_name = "myinstance"
  auth_token    = "your-evolution-api-token"
  recipient     = "+551198765432"
}
