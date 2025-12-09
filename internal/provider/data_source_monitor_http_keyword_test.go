package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMonitorHTTPKeywordDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPKeywordMonitor")
	url := "https://httpbin.org/html"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPKeywordDataSourceConfig(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_http_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
			{
				Config: testAccMonitorHTTPKeywordDataSourceConfigByID(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_http_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccMonitorHTTPKeywordDataSourceConfig(name, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_keyword" "test" {
  name    = %[1]q
  url     = %[2]q
  keyword = "html"
}

data "uptimekuma_monitor_http_keyword" "test" {
  name = uptimekuma_monitor_http_keyword.test.name
}
`, name, url)
}

func testAccMonitorHTTPKeywordDataSourceConfigByID(name, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_keyword" "test" {
  name    = %[1]q
  url     = %[2]q
  keyword = "html"
}

data "uptimekuma_monitor_http_keyword" "test" {
  id = uptimekuma_monitor_http_keyword.test.id
}
`, name, url)
}
