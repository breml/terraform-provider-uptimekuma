resource "uptimekuma_status_page" "example" {
  slug      = "example-status"
  title     = "Example Status Page"
  published = true
}

resource "uptimekuma_status_page_incident" "example" {
  status_page_slug = uptimekuma_status_page.example.slug
  title            = "Scheduled Maintenance"
  content          = "We will be performing scheduled maintenance on our services."
  style            = "info"
  pin              = true
}
