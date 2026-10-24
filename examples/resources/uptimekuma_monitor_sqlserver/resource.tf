resource "uptimekuma_monitor_sqlserver" "example" {
  name     = "SQL Server Production Database"
  active   = true
  interval = 60

  # SQL Server connection string with authentication
  # Important: Ensure proper TLS certificate validation in production
  database_connection_string = "Server=mssql.example.com;User=sa;Password=YourPassword123"

  # Optional SQL query for custom health check (default: SELECT 1)
  database_query = "SELECT 1"
}

resource "uptimekuma_monitor_sqlserver" "with_group" {
  name     = "SQL Server Grouped Monitor"
  active   = true
  interval = 60
  parent   = uptimekuma_monitor_group.databases.id

  # For development/testing with self-signed certificates, add TrustServerCertificate=true
  database_connection_string = "Server=mssql.example.com;User=sa;Password=YourPassword123"
  database_query             = "SELECT COUNT(*) FROM sys.databases"
}

resource "uptimekuma_monitor_group" "databases" {
  name = "Database Monitors"
}

# Requires Terraform or OpenTofu 1.11 or later: database_connection_string_wo is never stored in
# plan or state. Change database_connection_string_wo_version whenever the secret changes.
resource "uptimekuma_monitor_sqlserver" "write_only" {
  name = "SQL Server monitor with a write-only connection string"

  database_connection_string_wo         = "Server=sqlserver.example.com,1433;Database=app;User Id=monitor;Password=secret;"
  database_connection_string_wo_version = "1"
}
