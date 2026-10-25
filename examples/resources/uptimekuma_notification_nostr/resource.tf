resource "uptimekuma_notification_nostr" "example" {
  name = "Nostr Notification"

  # Sender's private key in Nostr format (nsec encoded)
  # This is a sensitive value and should be protected
  sender = "nsec1vl5dr5z69cydfy5kqsruqh84zkyv684jpu7v0unv2dl0aq2t5spe5nqq"

  # Newline-delimited list of recipient public keys (npub encoded)
  # These are public keys and can be shared openly
  recipients = "npub1qypt4l5elx7qjxapqvzc3gw7nj5zxwq5r5rzc5yqgj5j5j5j5j5j5j5j5j5j"

  # Newline-delimited list of Nostr relay URLs
  relays = "wss://relay.example.com\nwss://relay2.example.com"

  # Optional: Enable or disable the notification
  is_active = true

  # Optional: Mark as default notification
  is_default = false

  # Optional: Apply to existing monitors
  apply_existing = false
}

# Requires Terraform or OpenTofu 1.11 or later: sender_wo is never stored in plan
# or state. Change the matching *_wo_version whenever a secret changes.
resource "uptimekuma_notification_nostr" "write_only" {
  name       = "Nostr with write-only secret"
  recipients = "npub1qypt4l5elx7qjxapqvzc3gw7nj5zxwq5r5rzc5yqgj5j5j5j5j5j5j5j5j5j"
  relays     = "wss://relay.example.com"

  sender_wo         = "nsec1vl5dr5z69cydfy5kqsruqh84zkyv684jpu7v0unv2dl0aq2t5spe5nqq"
  sender_wo_version = "1"
}
