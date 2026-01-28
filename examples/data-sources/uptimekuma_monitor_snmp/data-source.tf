# SNMP Monitor Data Source Example

# Reference an existing SNMP monitor by ID
data "uptimekuma_monitor_snmp" "by_id" {
  id = 42
}

# Reference an existing SNMP monitor by name
data "uptimekuma_monitor_snmp" "by_name" {
  name = "Network Device SNMP Monitor"
}

# Use data source output in another resource
resource "uptimekuma_notification_webhook" "snmp_webhook" {
  name = "SNMP Monitor Notifications"
  url  = "https://example.com/notify"
}

# Create a notification association using data source
resource "uptimekuma_monitor_snmp" "monitored" {
  name             = "Production SNMP Device"
  hostname         = "prod-router.internal"
  snmp_version     = "2c"
  snmp_oid         = ".1.3.6.1.2.1.1.5.0"
  snmp_community   = "monitoring"
  notification_ids = [uptimekuma_notification_webhook.snmp_webhook.id]
}
