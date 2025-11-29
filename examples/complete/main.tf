terraform {
  required_providers {
    uptimekuma = {
      source  = "breml/uptimekuma"
      version = "~> 0.1"
    }
  }
}

provider "uptimekuma" {
  endpoint = var.uptimekuma_endpoint
  username = var.uptimekuma_username
  password = var.uptimekuma_password
}

# Monitor Groups
resource "uptimekuma_monitor_group" "production" {
  name   = "Production"
  status = "up"
  active = true
}

resource "uptimekuma_monitor_group" "staging" {
  name   = "Staging"
  status = "up"
  active = true
}

# Notification Channels
resource "uptimekuma_notification_slack" "critical_alerts" {
  name            = "Critical Alerts - Slack"
  webhook_url     = var.slack_webhook_url
  channel         = "#alerts"
  mention_channel = "channel"
  is_active       = true
  is_default      = true
}

resource "uptimekuma_notification_ntfy" "fallback" {
  name      = "Ntfy Fallback"
  topic     = "uptime_kuma_alerts"
  priority  = "high"
  server    = "https://ntfy.sh"
  is_active = true
}

# Production Monitors
resource "uptimekuma_monitor_http" "api_prod" {
  name           = "Production API"
  url            = "https://api.example.com/health"
  interval       = 60
  timeout        = 30
  max_retries    = 2
  retry_interval = 60
  upside_down    = false
  active         = true
  method         = "GET"
  ignore_tls     = false
  max_redirects  = 10
  parent         = uptimekuma_monitor_group.production.id
}

resource "uptimekuma_monitor_http_keyword" "status_page" {
  name     = "Status Page - All Green"
  url      = "https://status.example.com"
  keyword  = "All systems operational"
  interval = 300
  timeout  = 30
  active   = true
  parent   = uptimekuma_monitor_group.production.id
}

resource "uptimekuma_monitor_postgres" "database_prod" {
  name              = "Production Database"
  hostname          = var.db_host
  port              = 5432
  database_user     = var.db_user
  database_password = var.db_password
  database_name     = "production"
  interval          = 600
  timeout           = 30
  max_retries       = 2
  active            = true
  sql_query         = "SELECT 1"
  parent            = uptimekuma_monitor_group.production.id
}

resource "uptimekuma_monitor_ping" "server_ping" {
  name           = "Server Connectivity"
  hostname       = "prod.example.com"
  interval       = 120
  timeout        = 30
  max_retries    = 2
  retry_interval = 60
  upside_down    = false
  active         = true
  packet_size    = 56
  parent         = uptimekuma_monitor_group.production.id
}

# Staging Monitors
resource "uptimekuma_monitor_http" "api_staging" {
  name          = "Staging API"
  url           = "https://staging-api.example.com/health"
  interval      = 120
  timeout       = 30
  max_retries   = 1
  active        = true
  method        = "GET"
  ignore_tls    = false
  max_redirects = 10
  parent        = uptimekuma_monitor_group.staging.id
}

resource "uptimekuma_monitor_tcp_port" "cache_staging" {
  name        = "Cache Server - Staging"
  hostname    = "cache-staging.example.com"
  port        = 6379
  interval    = 60
  timeout     = 30
  max_retries = 1
  active      = true
  parent      = uptimekuma_monitor_group.staging.id
}

# External Push Monitor for CI/CD Integration
resource "uptimekuma_monitor_push" "deployment_monitor" {
  name        = "Deployment Events"
  interval    = 3600
  timeout     = 30
  upside_down = false
  active      = true
}
