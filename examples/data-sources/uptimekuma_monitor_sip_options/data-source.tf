# Example: Read SIP Options monitor by name
data "uptimekuma_monitor_sip_options" "example" {
  name = "SIP Server Monitoring"
}

# Example: Read SIP Options monitor by ID
data "uptimekuma_monitor_sip_options" "by_id" {
  id = 1
}
