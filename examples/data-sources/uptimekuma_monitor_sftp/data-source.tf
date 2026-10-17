# Look up an SFTP monitor by name
data "uptimekuma_monitor_sftp" "backup" {
  name = "Backup SFTP"
}

# Look up an SFTP monitor by ID
data "uptimekuma_monitor_sftp" "by_id" {
  id = 42
}

# Use the data source to reference an existing monitor. The SSH credentials are
# not exposed by the data source.
output "sftp_hostname" {
  value = data.uptimekuma_monitor_sftp.backup.hostname
}
