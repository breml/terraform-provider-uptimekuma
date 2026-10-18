# A monitor whose status is set from Terraform, for example to report a
# component that is checked by an external process
resource "uptimekuma_monitor_manual" "example" {
  name          = "Payment Provider"
  description   = "Status reported by the on-call team"
  manual_status = "up"
}

# Without manual_status the monitor reports pending
resource "uptimekuma_monitor_manual" "pending" {
  name     = "Legacy Backend"
  interval = 120
}
