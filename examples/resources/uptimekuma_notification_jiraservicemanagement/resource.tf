resource "uptimekuma_notification_jiraservicemanagement" "example" {
  name      = "Jira Service Management Example"
  is_active = true
  cloud_id  = "your-jsm-cloud-id"
  email     = "you@example.com"
  api_token = "your-jsm-api-token"
  priority  = 2
}
