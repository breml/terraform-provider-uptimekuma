# OracleDB monitor resource (EZCONNECT format)
resource "uptimekuma_monitor_oracledb" "example" {
  name                       = "Oracle Database Health Check"
  database_connection_string = "localhost:1521/ORCL"
  database_query             = "SELECT 1 FROM DUAL"
  basic_auth_user            = "monitoring_user"
  basic_auth_pass            = "secure_password"
  interval                   = 60
  retry_interval             = 60
  active                     = true
}

# OracleDB monitor with optional fields
resource "uptimekuma_monitor_oracledb" "example_advanced" {
  name                       = "Advanced Oracle Check"
  description                = "Monitors Oracle database health with custom query"
  database_connection_string = "db.example.com:1521/PRODDB"
  database_query             = "SELECT COUNT(*) FROM user_tables"
  basic_auth_user            = "app_user"
  basic_auth_pass            = "secret"
  interval                   = 30
  retry_interval             = 10
  resend_interval            = 0
  max_retries                = 5
  upside_down                = false
  active                     = true
}

# OracleDB monitor with tags
resource "uptimekuma_tag" "oracle_environment" {
  name = "environment"
}

resource "uptimekuma_tag" "oracle_service" {
  name = "service"
}

resource "uptimekuma_monitor_oracledb" "example_with_tags" {
  name                       = "Oracle with Tags"
  database_connection_string = "localhost:1521/ORCL"
  basic_auth_user            = "user"
  basic_auth_pass            = "pass"
  interval                   = 60
  active                     = true

  tags = [
    {
      tag_id = uptimekuma_tag.oracle_environment.id
      value  = "production"
    },
    {
      tag_id = uptimekuma_tag.oracle_service.id
      value  = "database"
    },
  ]
}

# OracleDB monitor in a group
resource "uptimekuma_monitor_group" "database_monitors" {
  name = "Database Monitors"
}

resource "uptimekuma_monitor_oracledb" "example_grouped" {
  name                       = "Oracle in Group"
  database_connection_string = "db-server:1521/MYDB"
  basic_auth_user            = "user"
  basic_auth_pass            = "password"
  interval                   = 60
  active                     = true
  parent                     = uptimekuma_monitor_group.database_monitors.id
}

# OracleDB monitor with notification
resource "uptimekuma_monitor_oracledb" "example_with_notification" {
  name                       = "Oracle with Alerts"
  database_connection_string = "localhost:1521/METRICS"
  basic_auth_user            = "monitoring"
  basic_auth_pass            = "pass"
  database_query             = "SELECT 1 FROM DUAL"
  interval                   = 60
  active                     = true

  notification_ids = [uptimekuma_notification_slack.alerts.id]
}
