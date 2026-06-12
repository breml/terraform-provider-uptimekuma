resource "uptimekuma_notification_webpush" "example" {
  name      = "Web Push Notifications"
  is_active = true

  subscription = {
    endpoint = "https://fcm.googleapis.com/fcm/send/abc123"
    keys = {
      p256dh = "BGxi5eHcCnFv1example..."
      auth   = "auth-secret-example"
    }
  }
}
