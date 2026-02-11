data "uptimekuma_notification_heiioncall" "example" {
  name = "Heii On-Call Alerts"
}

output "heii_oncall_id" {
  value       = data.uptimekuma_notification_heiioncall.example.id
  description = "The ID of the Heii On-Call notification"
}
