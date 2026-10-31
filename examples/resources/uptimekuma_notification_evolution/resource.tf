resource "uptimekuma_notification_evolution" "example" {
  name          = "Evolution WhatsApp Notification"
  api_url       = "https://api.evolution.example.com"
  instance_name = "myinstance"
  auth_token    = "your-evolution-api-token"
  recipient     = "+551198765432"
}

# Requires Terraform or OpenTofu 1.11 or later: auth_token_wo is never stored in
# plan or state. Change auth_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_evolution" "write_only" {
  name = "Evolution with a write-only secret"

  api_url       = "https://evolution.example.com"
  instance_name = "uptime"
  recipient     = "5511999999999"

  auth_token_wo         = "your-evolution-api-key"
  auth_token_wo_version = "1"
}
