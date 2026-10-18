# Example: Read manual monitor by name
data "uptimekuma_monitor_manual" "example" {
  name = "Payment Provider"
}

# Example: Read manual monitor by ID
data "uptimekuma_monitor_manual" "by_id" {
  id = 1
}
