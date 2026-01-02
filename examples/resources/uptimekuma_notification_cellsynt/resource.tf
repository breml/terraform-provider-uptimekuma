resource "uptimekuma_notification_cellsynt" "example" {
  name            = "Cellsynt SMS Notification"
  is_active       = true
  login           = "your_cellsynt_username"
  password        = "your_cellsynt_password"
  destination     = "+46701234567"
  originator      = "YourCompany"
  originator_type = "Numeric"
  allow_long_sms  = false
}
