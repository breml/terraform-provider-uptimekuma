output "production_group_id" {
  description = "Production monitor group ID"
  value       = uptimekuma_monitor_group.production.id
}

output "staging_group_id" {
  description = "Staging monitor group ID"
  value       = uptimekuma_monitor_group.staging.id
}

output "api_prod_monitor_id" {
  description = "Production API monitor ID"
  value       = uptimekuma_monitor_http.api_prod.id
}

output "deployment_push_monitor_url" {
  description = "Deployment push monitor URL for external integrations"
  value       = uptimekuma_monitor_push.deployment_monitor.push_token
  sensitive   = true
}

output "slack_notification_id" {
  description = "Slack notification channel ID"
  value       = uptimekuma_notification_slack.critical_alerts.id
}
