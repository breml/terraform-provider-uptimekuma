# SFTP Monitor Resource Example

# Password authentication. Leaving port unset stores it as SQL NULL, which tells
# the check to fall back to port 22; setting port = 22 is a different state, so
# prefer leaving it out.
resource "uptimekuma_monitor_sftp" "password" {
  name         = "Backup SFTP"
  hostname     = "sftp.example.com"
  ssh_username = "backup"
  ssh_password = var.sftp_password
}

# Private key authentication, additionally checking that a remote path exists
resource "uptimekuma_monitor_sftp" "private_key" {
  name        = "Deploy SFTP"
  description = "Upload target of the release pipeline"
  hostname    = "files.example.com"
  port        = 2222
  timeout     = 15

  ssh_username    = "deploy"
  ssh_auth_method = "privateKey"
  ssh_private_key = file("${path.module}/id_ed25519")
  ssh_passphrase  = var.sftp_key_passphrase

  # While set, the check also verifies that this path exists on the server.
  sftp_path = "/srv/releases"

  interval        = 300
  retry_interval  = 120
  resend_interval = 10
  max_retries     = 5
  active          = true
  upside_down     = false

  notification_ids = [1, 2]

  tags = [
    {
      tag_id = 1
      value  = "production"
    }
  ]
}

variable "sftp_password" {
  type      = string
  sensitive = true
}

variable "sftp_key_passphrase" {
  type      = string
  sensitive = true
}
