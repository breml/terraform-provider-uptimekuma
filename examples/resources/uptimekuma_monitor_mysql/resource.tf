# MySQL monitor resource
resource "uptimekuma_monitor_mysql" "example" {
  name                       = "MySQL Database Health Check"
  database_connection_string = "root:password@tcp(localhost:3306)/mydb"
  database_query             = "SELECT 1"
  interval                   = 60
  retry_interval             = 60
  active                     = true
}

# MySQL monitor with optional fields
resource "uptimekuma_monitor_mysql" "example_advanced" {
  name                       = "Advanced MySQL Check"
  description                = "Monitors MySQL database health with custom query"
  database_connection_string = "app_user:secure_password@tcp(db.example.com:3306)/production"
  database_query             = "SELECT VERSION()"
  interval                   = 30
  retry_interval             = 10
  resend_interval            = 0
  max_retries                = 5
  upside_down                = false
  active                     = true
}

# MySQL monitor with tags
resource "uptimekuma_tag" "mysql_environment" {
  name = "environment"
}

resource "uptimekuma_tag" "mysql_service" {
  name = "service"
}

resource "uptimekuma_monitor_mysql" "example_with_tags" {
  name                       = "MySQL with Tags"
  database_connection_string = "user:pass@tcp(localhost:3306)/db"
  database_query             = "SELECT COUNT(*) FROM users"
  interval                   = 60
  active                     = true

  tags = [
    {
      tag_id = uptimekuma_tag.mysql_environment.id
      value  = "production"
    },
    {
      tag_id = uptimekuma_tag.mysql_service.id
      value  = "database"
    },
  ]
}

# MySQL monitor in a group
resource "uptimekuma_monitor_group" "database_monitors" {
  name = "Database Monitors"
}

resource "uptimekuma_monitor_mysql" "example_grouped" {
  name                       = "MySQL in Group"
  database_connection_string = "user:password@tcp(db-server:3306)/mydb"
  interval                   = 60
  active                     = true
  parent                     = uptimekuma_monitor_group.database_monitors.id
}

# MySQL monitor with notification
resource "uptimekuma_monitor_mysql" "example_with_notification" {
  name                       = "MySQL with Alerts"
  database_connection_string = "monitoring:pass@tcp(localhost:3306)/metrics"
  database_query             = "SELECT 1"
  interval                   = 60
  active                     = true

  notification_ids = [uptimekuma_notification_slack.alerts.id]
}
