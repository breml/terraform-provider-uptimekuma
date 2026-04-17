resource "uptimekuma_notification_whapi" "example" {
  name       = "Whapi Notifications"
  api_url    = "https://gate.whapi.cloud"
  auth_token = "your-whapi-auth-token"
  recipient  = "1234567890"
  is_active  = true
  is_default = false
}
