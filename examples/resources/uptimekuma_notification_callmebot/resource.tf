resource "uptimekuma_notification_callmebot" "example" {
  name       = "CallMeBot WhatsApp"
  endpoint   = "https://api.callmebot.com/whatsapp.php?phone=YOUR_PHONE&text="
  is_default = false
}

resource "uptimekuma_notification_callmebot" "telegram_example" {
  name       = "CallMeBot Telegram"
  endpoint   = "https://api.callmebot.com/telegram.php?token=YOUR_TOKEN&chat_id=YOUR_CHAT_ID&text="
  is_default = false
}

# Requires Terraform or OpenTofu 1.11 or later: endpoint_wo is never stored in
# plan or state. Change endpoint_wo_version whenever the secret changes.
resource "uptimekuma_notification_callmebot" "write_only" {
  name = "CallMeBot with a write-only secret"

  endpoint_wo         = "https://api.callmebot.com/whatsapp.php?phone=+41790000000&apikey=your-api-key"
  endpoint_wo_version = "1"
}
