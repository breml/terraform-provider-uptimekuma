resource "uptimekuma_monitor_real_browser" "example" {
  name              = "Browser-based Synthetic Monitoring"
  url               = "https://example.com"
  interval          = 300
  timeout           = 60
  max_retries       = 2
  upside_down       = false
  active            = true
  automation_script = "step('load page', async () => { await page.goto('https://example.com'); });"
}
