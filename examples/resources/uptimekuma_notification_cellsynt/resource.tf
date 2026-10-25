resource "uptimekuma_notification_cellsynt" "example" {
  name            = "Cellsynt SMS Notification"
  is_active       = true
  login           = "your_cellsynt_username"
  password        = "your_cellsynt_password"
  destination     = "+46701234567"
  originator      = "YourCompany"
  originator_type = "Alphanumeric"
  allow_long_sms  = false
}

# Requires Terraform or OpenTofu 1.11 or later: password_wo is never stored in
# plan or state. Change password_wo_version whenever the secret changes.
resource "uptimekuma_notification_cellsynt" "write_only" {
  name = "Cellsynt with a write-only secret"

  login       = "cellsynt-user"
  destination = "0046700000000"
  originator  = "UptimeKuma"

  password_wo         = "your-cellsynt-password"
  password_wo_version = "1"
}
