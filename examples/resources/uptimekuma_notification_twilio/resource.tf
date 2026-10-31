resource "uptimekuma_notification_twilio" "example" {
  name        = "Twilio SMS Notifications"
  account_sid = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  auth_token  = "your-auth-token"
  to_number   = "+12025550123"
  from_number = "+12025550456"
  is_default  = false
}

resource "uptimekuma_notification_twilio" "with_api_key" {
  name        = "Twilio with API Key"
  account_sid = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  api_key     = "your-api-key"
  auth_token  = "your-auth-token"
  to_number   = "+12025550123"
  from_number = "+12025550456"
  is_default  = false
}

# Requires Terraform or OpenTofu 1.11 or later: api_key_wo and auth_token_wo are
# never stored in plan or state. Change the matching *_wo_version whenever a
# secret changes.
resource "uptimekuma_notification_twilio" "write_only" {
  name        = "Twilio with write-only secrets"
  account_sid = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  to_number   = "+12025550123"
  from_number = "+12025550456"

  api_key_wo            = "your-api-key"
  api_key_wo_version    = "1"
  auth_token_wo         = "your-auth-token"
  auth_token_wo_version = "1"
}
