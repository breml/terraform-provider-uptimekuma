resource "uptimekuma_monitor_redis" "example" {
  name              = "Redis Database Monitoring"
  hostname          = "redis.example.com"
  port              = 6379
  database_user     = ""
  database_password = "redis_password"
  database_index    = 0
  interval          = 60
  timeout           = 30
  max_retries       = 2
  upside_down       = false
  active            = true
}

# Requires Terraform or OpenTofu 1.11 or later: database_connection_string_wo is never stored in
# plan or state. Change database_connection_string_wo_version whenever the secret changes.
resource "uptimekuma_monitor_redis" "write_only" {
  name = "Redis monitor with a write-only connection string"

  database_connection_string_wo         = "redis://monitor:secret@redis.example.com:6379"
  database_connection_string_wo_version = "1"
}
