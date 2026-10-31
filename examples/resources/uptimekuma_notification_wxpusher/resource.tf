resource "uptimekuma_notification_wxpusher" "example" {
  name       = "WxPusher Notifications"
  spt        = "SPT_0123456789ab"
  is_default = false
}

# Deliver to several recipients by separating their tokens with commas
resource "uptimekuma_notification_wxpusher" "team" {
  name       = "WxPusher Team Notifications"
  spt        = "SPT_0123456789ab,SPT_cdef01234567"
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: spt_wo is never stored in
# plan or state. Change spt_wo_version whenever the secret changes.
resource "uptimekuma_notification_wxpusher" "write_only" {
  name = "WxPusher with a write-only secret"

  spt_wo         = "SPT_0123456789ab"
  spt_wo_version = "1"
}
