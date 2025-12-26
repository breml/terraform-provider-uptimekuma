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

func TestAccMonitorHTTPJSONQueryDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPJSONQueryMonitor")
	url := "https://httpbin.org/json"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPJSONQueryDataSourceConfig(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http_json_query.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorHTTPJSONQueryDataSourceConfigByID(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http_json_query.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorHTTPJSONQueryDataSourceConfig(name, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_json_query" "test" {
  name       = %[1]q
  url        = %[2]q
  json_path  = "$.slideshow"
  expected_value = ""
}

data "uptimekuma_monitor_http_json_query" "test" {
  name = uptimekuma_monitor_http_json_query.test.name
}
`, name, url)
}

func testAccMonitorHTTPJSONQueryDataSourceConfigByID(name, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_json_query" "test" {
  name       = %[1]q
  url        = %[2]q
  json_path  = "$.slideshow"
  expected_value = ""
}

data "uptimekuma_monitor_http_json_query" "test" {
  id = uptimekuma_monitor_http_json_query.test.id
}
`, name, url)
}
