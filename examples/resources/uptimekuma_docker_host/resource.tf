# Example 1: Docker host using Unix socket (local)
resource "uptimekuma_docker_host" "local_socket" {
  name          = "Local Docker"
  docker_type   = "socket"
  docker_daemon = "unix:///var/run/docker.sock"
}

# Example 2: Docker host using TCP connection (remote)
resource "uptimekuma_docker_host" "remote_tcp" {
  name          = "Remote Docker Host"
  docker_type   = "tcp"
  docker_daemon = "tcp://docker.example.com:2375"
}

# Example 3: Docker host with TLS (secure remote connection)
resource "uptimekuma_docker_host" "secure_remote" {
  name          = "Secure Remote Docker"
  docker_type   = "tcp"
  docker_daemon = "tcp://secure-docker.example.com:2376"
}
