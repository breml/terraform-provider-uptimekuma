# Look up a proxy by ID
data "uptimekuma_proxy" "corporate" {
  id = 1
}

# Output proxy details
output "proxy_endpoint" {
  value = "${data.uptimekuma_proxy.corporate.protocol}://${data.uptimekuma_proxy.corporate.host}:${data.uptimekuma_proxy.corporate.port}"
}
