resource "uptimekuma_notification_wxpusher" "example" {
  name       = "WxPusher Notifications"
  spt        = "SPT_0123456789ab"
  is_active  = true
  is_default = false
}

# Deliver to several recipients by separating their tokens with commas
resource "uptimekuma_notification_wxpusher" "team" {
  name       = "WxPusher Team Notifications"
  spt        = "SPT_0123456789ab,SPT_cdef01234567"
  is_active  = true
  is_default = false
}
