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
