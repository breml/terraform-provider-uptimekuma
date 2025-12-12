# Example 1: Basic tag for production environment
resource "uptimekuma_tag" "production" {
  name  = "production"
  color = "#00ff00"
}

# Example 2: Tag for staging environment
resource "uptimekuma_tag" "staging" {
  name  = "staging"
  color = "#ffa500"
}

# Example 3: Tag for API services
resource "uptimekuma_tag" "api" {
  name  = "api"
  color = "#0066cc"
}

# Example 4: Using tags with a monitor
resource "uptimekuma_tag" "critical" {
  name  = "critical"
  color = "#ff0000"
}

resource "uptimekuma_monitor_http" "api_endpoint" {
  name = "Critical API Endpoint"
  url  = "https://api.example.com/health"

  tags = [
    {
      tag_id = uptimekuma_tag.production.id
    },
    {
      tag_id = uptimekuma_tag.critical.id
    },
    {
      tag_id = uptimekuma_tag.api.id
      value  = "v2"
    },
  ]
}
