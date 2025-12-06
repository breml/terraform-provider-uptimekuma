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

func TestAccMonitorRealBrowserResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRealBrowserMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestRealBrowserMonitorUpdated")
	url := "https://example.com"
	urlUpdated := "https://example.org"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorRealBrowserResourceConfig(name, url, 60, 48),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(48)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorRealBrowserResourceConfig(nameUpdated, urlUpdated, 120, 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("url"), knownvalue.StringExact(urlUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccMonitorRealBrowserResourceConfig(name, url string, interval, timeout int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name     = %[1]q
  url      = %[2]q
  interval = %[3]d
  timeout  = %[4]d
  active   = true
}
`, name, url, interval, timeout)
}

func TestAccMonitorRealBrowserResourceWithStatusCodes(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRealBrowserMonitorWithStatusCodes")
	url := "https://example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRealBrowserResourceConfigWithStatusCodes(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_real_browser.test", tfjsonpath.New("accepted_status_codes"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("200-299"),
							knownvalue.StringExact("301"),
						})),
				},
			},
		},
	})
}

func testAccMonitorRealBrowserResourceConfigWithStatusCodes(name, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name                  = %[1]q
  url                   = %[2]q
  accepted_status_codes = ["200-299", "301"]
}
`, name, url)
}

func TestAccMonitorRealBrowserResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRealBrowserMonitorImport")
	url := "https://example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRealBrowserResourceConfig(name, url, 60, 48),
			},
			{
				ResourceName:      "uptimekuma_monitor_real_browser.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
