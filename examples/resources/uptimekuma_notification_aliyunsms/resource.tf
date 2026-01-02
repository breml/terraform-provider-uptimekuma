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
