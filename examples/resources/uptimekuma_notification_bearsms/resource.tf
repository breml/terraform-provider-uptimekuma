resource "uptimekuma_notification_bearsms" "example" {
  name         = "BearSMS Notifications"
  username     = "your-bearsms-username"
  hash_key     = "your-bearsms-hash-key"
  phone_number = "9725XXXXXXXX"
  is_active    = true
  is_default   = false
}

# Show an approved sender name to the recipient (at most 11 characters).
resource "uptimekuma_notification_bearsms" "with_sender" {
  name         = "BearSMS with sender"
  username     = "your-bearsms-username"
  hash_key     = "your-bearsms-hash-key"
  sender_id    = "UptimeKuma"
  phone_number = "9725XXXXXXXX"
}

# Requires Terraform or OpenTofu 1.11 or later: hash_key_wo is never stored in
# plan or state. Change hash_key_wo_version whenever the hash key changes.
resource "uptimekuma_notification_bearsms" "write_only" {
  name = "BearSMS with a write-only hash key"

  username     = "your-bearsms-username"
  phone_number = "9725XXXXXXXX"

  hash_key_wo         = "your-bearsms-hash-key"
  hash_key_wo_version = "1"
}
