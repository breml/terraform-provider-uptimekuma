# MongoDB monitor resource
resource "uptimekuma_monitor_mongodb" "example" {
  name                       = "MongoDB Database Health Check"
  database_connection_string = "mongodb://user:password@localhost:27017/mydb"
  database_query             = "{\"ping\": 1}"
  interval                   = 60
  retry_interval             = 60
  active                     = true
}

# MongoDB monitor with optional fields
resource "uptimekuma_monitor_mongodb" "example_advanced" {
  name                       = "Advanced MongoDB Check"
  description                = "Monitors MongoDB database health with custom command"
  database_connection_string = "mongodb://monitoring_user:secure_password@db.example.com:27017/production"
  database_query             = "{\"serverStatus\": 1}"
  interval                   = 30
  retry_interval             = 10
  resend_interval            = 0
  max_retries                = 5
  upside_down                = false
  active                     = true
}

# MongoDB monitor with JSON path validation
resource "uptimekuma_monitor_mongodb" "example_with_json_path" {
  name                       = "MongoDB with JSON Path Validation"
  database_connection_string = "mongodb://user:password@localhost:27017/testdb"
  database_query             = "{\"find\": \"test_collection\"}"
  json_path                  = "$.ok"
  expected_value             = "1"
  interval                   = 60
  active                     = true
}

# MongoDB monitor with tags
resource "uptimekuma_tag" "mongodb_environment" {
  name = "environment"
}

resource "uptimekuma_tag" "mongodb_service" {
  name = "service"
}

resource "uptimekuma_monitor_mongodb" "example_with_tags" {
  name                       = "MongoDB with Tags"
  database_connection_string = "mongodb://user:password@localhost:27017/db"
  database_query             = "{\"ping\": 1}"
  interval                   = 60
  active                     = true

  tags = [
    {
      tag_id = uptimekuma_tag.mongodb_environment.id
      value  = "production"
    },
    {
      tag_id = uptimekuma_tag.mongodb_service.id
      value  = "database"
    },
  ]
}

# MongoDB monitor in a group
resource "uptimekuma_monitor_group" "database_monitors" {
  name = "Database Monitors"
}

resource "uptimekuma_monitor_mongodb" "example_grouped" {
  name                       = "MongoDB in Group"
  database_connection_string = "mongodb://user:password@db-server:27017/mydb"
  interval                   = 60
  active                     = true
  parent                     = uptimekuma_monitor_group.database_monitors.id
}

# MongoDB monitor with notification
resource "uptimekuma_monitor_mongodb" "example_with_notification" {
  name                       = "MongoDB with Alerts"
  database_connection_string = "mongodb://monitoring:password@localhost:27017/metrics"
  database_query             = "{\"ping\": 1}"
  interval                   = 60
  active                     = true

  notification_ids = [uptimekuma_notification_slack.alerts.id]
}
