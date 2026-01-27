resource "uptimekuma_monitor_sqlserver" "example" {
  name     = "SQL Server Production Database"
  active   = true
  interval = 60

  # SQL Server connection string with authentication
  database_connection_string = "Server=mssql.example.com;User=sa;Password=YourPassword123;TrustServerCertificate=true"

  # Optional SQL query for custom health check (default: SELECT 1)
  database_query = "SELECT 1"
}

resource "uptimekuma_monitor_sqlserver" "with_group" {
  name     = "SQL Server Grouped Monitor"
  active   = true
  interval = 60
  parent   = uptimekuma_monitor_group.databases.id

  database_connection_string = "Server=mssql.example.com;User=sa;Password=YourPassword123;TrustServerCertificate=true"
  database_query             = "SELECT COUNT(*) FROM sys.databases"
}

resource "uptimekuma_monitor_group" "databases" {
  name = "Database Monitors"
}
