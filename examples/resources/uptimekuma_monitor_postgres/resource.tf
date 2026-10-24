resource "uptimekuma_monitor_postgres" "example" {
  name              = "PostgreSQL Database Monitoring"
  hostname          = "db.example.com"
  port              = 5432
  database_user     = "monitoring_user"
  database_password = "secure_password"
  database_name     = "myapp"
  interval          = 300
  timeout           = 30
  max_retries       = 2
  upside_down       = false
  active            = true
  sql_query         = "SELECT 1"
}

# Requires Terraform or OpenTofu 1.11 or later: database_connection_string_wo is never stored in
# plan or state. Change database_connection_string_wo_version whenever the secret changes.
resource "uptimekuma_monitor_postgres" "write_only" {
  name = "PostgreSQL monitor with a write-only connection string"

  database_connection_string_wo         = "postgres://monitor:secret@postgres.example.com:5432/app"
  database_connection_string_wo_version = "1"
}
