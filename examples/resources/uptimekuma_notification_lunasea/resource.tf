resource "uptimekuma_notification_lunasea" "user_target" {
  name            = "Lunasea User Notifications"
  target          = "user"
  lunasea_user_id = "YOUR_LUNASEA_USER_ID"
  is_active       = true
  is_default      = false
}

resource "uptimekuma_notification_lunasea" "device_target" {
  name       = "Lunasea Device Notifications"
  target     = "device"
  device     = "YOUR_LUNASEA_DEVICE_ID"
  is_active  = true
  is_default = false
}
