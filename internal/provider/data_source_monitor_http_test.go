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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPDataSourceConfig(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.by_name",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.by_name",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.by_id",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_http.by_id",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorHTTPDataSourceConfig(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name                       = %[1]q
  url                        = %[2]q
  domain_expiry_notification = true
}

data "uptimekuma_monitor_http" "by_name" {
  name = uptimekuma_monitor_http.test.name
}

data "uptimekuma_monitor_http" "by_id" {
  id = uptimekuma_monitor_http.test.id
}
`, name, url)
}
