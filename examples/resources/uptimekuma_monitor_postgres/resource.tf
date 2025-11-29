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
