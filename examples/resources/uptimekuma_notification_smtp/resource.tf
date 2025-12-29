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

  is_active  = true
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
  dkim_domain           = "example.com"
  dkim_key_selector     = "default"
  dkim_private_key      = "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
  dkim_hash_algo        = "sha256"
  dkim_header_field_names = "From:Subject:Date"
  dkim_skip_fields      = "Bcc"

  is_active  = true
  is_default = false
}
