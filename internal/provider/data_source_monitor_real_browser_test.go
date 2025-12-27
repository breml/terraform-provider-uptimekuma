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

func TestAccMonitorRealBrowserDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRealBrowserMonitor")
	url := "https://httpbin.org/status/200"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRealBrowserDataSourceConfig(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorRealBrowserDataSourceConfigByID(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorRealBrowserDataSourceConfig(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name = %[1]q
  url  = %[2]q
}

data "uptimekuma_monitor_real_browser" "test" {
  name = uptimekuma_monitor_real_browser.test.name
}
`, name, url)
}

func testAccMonitorRealBrowserDataSourceConfigByID(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name = %[1]q
  url  = %[2]q
}

data "uptimekuma_monitor_real_browser" "test" {
  id = uptimekuma_monitor_real_browser.test.id
}
`, name, url)
}
