# Look up a tag by name
data "uptimekuma_tag" "environment" {
  name = "production"
}

# Look up by ID
data "uptimekuma_tag" "by_id" {
  id = 5
}

# Use the tag with a monitor
resource "uptimekuma_monitor_http" "api" {
  name = "API Service"
  url  = "https://api.example.com"

  tags = [
    {
      tag_id = data.uptimekuma_tag.environment.id
      value  = "prod"
    }
  ]
}
