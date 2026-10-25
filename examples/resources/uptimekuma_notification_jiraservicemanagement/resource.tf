resource "uptimekuma_notification_jiraservicemanagement" "example" {
  name      = "Jira Service Management Example"
  is_active = true
  cloud_id  = "your-jsm-cloud-id"
  email     = "you@example.com"
  api_token = "your-jsm-api-token"
  priority  = 2
}

# Requires Terraform or OpenTofu 1.11 or later: api_token_wo is never stored in
# plan or state. Change api_token_wo_version whenever the secret changes.
resource "uptimekuma_notification_jiraservicemanagement" "write_only" {
  name     = "Jira Service Management with write-only token"
  cloud_id = "your-jsm-cloud-id"
  email    = "you@example.com"

  api_token_wo         = "your-jsm-api-token"
  api_token_wo_version = "1"
}
