# SNMP Monitor Resource Example

# Basic SNMP monitor with minimum required configuration
resource "uptimekuma_monitor_snmp" "basic" {
  name           = "Network Device SNMP Monitor"
  hostname       = "192.168.1.1"
  snmp_version   = "2c"
  snmp_oid       = ".1.3.6.1.2.1.1.5.0"
  snmp_community = "public"
}

# SNMP monitor with all available options
resource "uptimekuma_monitor_snmp" "full" {
  name               = "Network Device SNMP Monitor Full"
  description        = "Monitor network device SNMP parameters"
  hostname           = "192.168.1.1"
  port               = 161
  snmp_version       = "2c"
  snmp_oid           = ".1.3.6.1.2.1.1.5.0"
  snmp_community     = "public"
  interval           = 120
  retry_interval     = 60
  max_retries        = 5
  active             = true
  upside_down        = false
  json_path          = "$.value"
  json_path_operator = "=="
  expected_value     = "1"

  notification_ids = [1, 2]

  tags = [
    {
      tag_id = 1
      value  = "production"
    }
  ]
}

# SNMP monitor using basic SNMPv3 configuration
resource "uptimekuma_monitor_snmp" "snmpv3" {
  name           = "SNMPv3 Device Monitor"
  hostname       = "192.168.1.2"
  snmp_version   = "3"
  snmp_oid       = ".1.3.6.1.2.1.1.3.0"
  snmp_community = "private"
  port           = 161
  interval       = 300
  active         = true
}

# SNMP monitor monitoring specific system uptime
resource "uptimekuma_monitor_snmp" "uptime" {
  name           = "Device Uptime Monitor"
  description    = "Monitor device system uptime via SNMP"
  hostname       = "192.168.1.3"
  snmp_version   = "2c"
  snmp_oid       = ".1.3.6.1.2.1.1.3.0" # sysUpTime.0
  snmp_community = "public"
  port           = 161
  interval       = 60
}
