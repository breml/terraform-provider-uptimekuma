# SMTP monitor resource for basic connectivity checking.
# This monitor checks connectivity to an SMTP mail server.
resource "uptimekuma_monitor_smtp" "mail_server" {
  name          = "Mail Server"
  description   = "Monitor SMTP mail server connectivity"
  hostname      = "smtp.example.com"
  port          = 587
  smtp_security = "STARTTLS"
  interval      = 60
}

# SMTP monitor with TLS security.
# Connects to an SMTP server using implicit TLS on port 465.
resource "uptimekuma_monitor_smtp" "secure_mail" {
  name          = "Secure Mail Server"
  hostname      = "mail.example.com"
  port          = 465
  smtp_security = "TLS"
  interval      = 30
  max_retries   = 3
}

# SMTP monitor without encryption.
# Checks plain SMTP connectivity on port 25.
resource "uptimekuma_monitor_smtp" "relay" {
  name          = "SMTP Relay"
  hostname      = "relay.internal.example.com"
  port          = 25
  smtp_security = "None"
  interval      = 120
  active        = true
}

# SMTP monitor as part of a monitor group.
# Organizes related monitors under a parent group.
resource "uptimekuma_monitor_smtp" "grouped_monitor" {
  name     = "Grouped SMTP Monitor"
  hostname = "smtp.example.com"
  parent   = uptimekuma_monitor_group.mail_monitors.id
  interval = 60
  active   = true
}

resource "uptimekuma_monitor_group" "mail_monitors" {
  name     = "Mail Monitors"
  interval = 60
}
