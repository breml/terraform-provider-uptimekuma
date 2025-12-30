resource "uptimekuma_notification_callmebot" "example" {
  name     = "CallMeBot WhatsApp"
  endpoint = "https://api.callmebot.com/whatsapp.php?phone=YOUR_PHONE&text="
  is_active   = true
  is_default  = false
}

resource "uptimekuma_notification_callmebot" "telegram_example" {
  name     = "CallMeBot Telegram"
  endpoint = "https://api.callmebot.com/telegram.php?token=YOUR_TOKEN&chat_id=YOUR_CHAT_ID&text="
  is_active   = true
  is_default  = false
}
