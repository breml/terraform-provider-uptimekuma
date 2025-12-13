# Look up an HTTP monitor by name
data "uptimekuma_monitor_http" "website" {
  name = "My Website"
}

# Look up an HTTP monitor by ID
data "uptimekuma_monitor_http" "by_id" {
  id = 42
}

# Use the data source to reference an existing monitor in other resources
output "website_url" {
  value = data.uptimekuma_monitor_http.website.url
}
