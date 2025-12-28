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

func TestAccMonitorHTTPDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitor")
	url := "https://httpbin.org/status/200"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPDataSourceConfig(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
				},
			},
			{
				Config: testAccMonitorHTTPDataSourceConfigByID(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
				},
			},
		},
	})
}

func testAccMonitorHTTPDataSourceConfig(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name = %[1]q
  url  = %[2]q
}

data "uptimekuma_monitor_http" "test" {
  name = uptimekuma_monitor_http.test.name
}
`, name, url)
}

func testAccMonitorHTTPDataSourceConfigByID(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name = %[1]q
  url  = %[2]q
}

data "uptimekuma_monitor_http" "test" {
  id = uptimekuma_monitor_http.test.id
}
`, name, url)
}
