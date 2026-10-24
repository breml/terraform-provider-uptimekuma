resource "uptimekuma_notification_aliyunsms" "example" {
  name              = "Aliyun SMS Notification"
  is_active         = true
  is_default        = false
  access_key_id     = "your-access-key-id"
  secret_access_key = "your-secret-access-key"
  phone_number      = "+1234567890"
  sign_name         = "YourCompanyName"
  template_code     = "SMS_001"
}

# Requires Terraform or OpenTofu 1.11 or later: secret_access_key_wo is never stored in
# plan or state. Change secret_access_key_wo_version whenever the secret changes.
resource "uptimekuma_notification_aliyunsms" "write_only" {
  name = "Aliyun SMS with a write-only secret"

  access_key_id = "LTAI0000"
  phone_number  = "13800000000"
  sign_name     = "UptimeKuma"
  template_code = "SMS_000000"

  secret_access_key_wo         = "your-secret-access-key"
  secret_access_key_wo_version = "1"
}
