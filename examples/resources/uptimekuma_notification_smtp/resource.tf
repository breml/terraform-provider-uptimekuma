resource "uptimekuma_notification_smtp" "example" {
  name = "SMTP Notifications"

  # Required fields
  host = "smtp.example.com"
  from = "uptime-kuma@example.com"
  to   = "admin@example.com"

  # Optional: SMTP server configuration
  port             = 587
  secure           = true
  ignore_tls_error = false

  # Optional: SMTP authentication
  username = "smtp-user"
  password = "smtp-password"

  # Optional: Additional recipients
  cc  = "supervisor@example.com"
  bcc = "archive@example.com"

  # Optional: Email customization
  custom_subject = "Alert from Uptime Kuma"
  custom_body    = "Service {{ serviceName }} is {{ status }}"
  html_body      = false

  # Optional: Additional mail headers
  additional_headers = jsonencode({
    "X-Custom-Header" = "Additional Header"
    "X-Environment"   = "production"
  })

  is_default = false
}

resource "uptimekuma_notification_smtp" "with_dkim" {
  name = "SMTP with DKIM"

  host = "smtp.example.com"
  port = 587
  from = "monitoring@example.com"
  to   = "alerts@example.com"

  username = "smtp-user"
  password = "smtp-password"

  # Optional: DKIM signing for authentication
  dkim_domain             = "example.com"
  dkim_key_selector       = "default"
  dkim_private_key        = "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
  dkim_hash_algo          = "sha256"
  dkim_header_field_names = "From:Subject:Date"
  dkim_skip_fields        = "Bcc"

  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: password_wo and
# dkim_private_key_wo are never stored in plan or state. Change the matching
# *_wo_version whenever a secret changes.
resource "uptimekuma_notification_smtp" "write_only_password" {
  name = "SMTP with write-only secrets"

  host = "smtp.example.com"
  port = 587
  from = "monitoring@example.com"
  to   = "alerts@example.com"

  username            = "smtp-user"
  password_wo         = "smtp-password"
  password_wo_version = "1"

  dkim_domain                 = "example.com"
  dkim_key_selector           = "default"
  dkim_private_key_wo         = "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
  dkim_private_key_wo_version = "1"

  is_default = false
}
