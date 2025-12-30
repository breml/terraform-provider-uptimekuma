# Example: Docker monitor for monitoring a Docker container
resource "uptimekuma_docker_host" "example" {
  name          = "My Docker Host"
  docker_daemon = "unix:///var/run/docker.sock"
  docker_type   = "socket"
}

resource "uptimekuma_monitor_docker" "example" {
  name             = "Redis Container"
  description      = "Monitor Redis container status"
  docker_host_id   = uptimekuma_docker_host.example.id
  docker_container = "redis"
  interval         = 60
  active           = true
}

# Example: Docker monitor with notifications
resource "uptimekuma_monitor_docker" "with_notifications" {
  name             = "App Container"
  docker_host_id   = uptimekuma_docker_host.example.id
  docker_container = "app"
  interval         = 60
  notification_ids = [1, 2]
  active           = true
}
