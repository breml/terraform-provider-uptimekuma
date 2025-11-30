resource "uptimekuma_monitor_http" "example" {
  name     = "Example Website"
  url      = "https://example.com"
  interval = 60
}

resource "uptimekuma_status_page" "example" {
  slug        = "example-status"
  title       = "Example Status Page"
  description = "Status page for our services"
  published   = true
  show_tags   = true
  theme       = "light"

  public_group_list = [
    {
      name   = "Production Services"
      weight = 1
      monitor_list = [
        {
          id       = uptimekuma_monitor_http.example.id
          send_url = false
        }
      ]
    }
  ]
}
