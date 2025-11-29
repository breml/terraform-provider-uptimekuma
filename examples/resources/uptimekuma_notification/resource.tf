resource "uptimekuma_notification" "example" {
  name = "Generic Custom Notification"
  type = "custom"
  config = jsonencode({
    customField = "customValue"
  })
  is_active      = true
  is_default     = false
  apply_existing = false
}
