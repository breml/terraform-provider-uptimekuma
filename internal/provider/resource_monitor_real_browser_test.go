package provider

import (
	"fmt"
	"regexp"
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
				Config:             testAccMonitorRealBrowserResourceConfig(name, url, 60, 48, true),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(48),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorRealBrowserResourceConfig(nameUpdated, urlUpdated, 120, 60, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorRealBrowserResourceConfig(
	name string, url string, interval int64, timeout float64, domainExpiry bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name                        = %[1]q
  url                         = %[2]q
  interval                    = %[3]d
  timeout                     = %[4]v
  active                      = true
  domain_expiry_notification  = %[5]t
}
`, name, url, interval, timeout, domainExpiry)
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
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("accepted_status_codes"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("200-299"),
							knownvalue.StringExact("301"),
						}),
					),
				},
			},
		},
	})
}

func testAccMonitorRealBrowserResourceConfigWithStatusCodes(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name                  = %[1]q
  url                   = %[2]q
  accepted_status_codes = ["200-299", "301"]
}
`, name, url)
}

func TestAccMonitorRealBrowserResourceWithScreenshotDelay(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRealBrowserScreenshotDelay")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRealBrowserResourceConfigScreenshotDelay(name, 2000),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("screenshot_delay"),
						knownvalue.Int64Exact(2000),
					),
				},
			},
			{
				Config: testAccMonitorRealBrowserResourceConfigScreenshotDelay(name, 5000),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("screenshot_delay"),
						knownvalue.Int64Exact(5000),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_real_browser.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorRealBrowserResourceConfigScreenshotDelay(name string, screenshotDelay int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name             = %[1]q
  url              = "https://example.com"
  screenshot_delay = %[2]d
}
`, name, screenshotDelay)
}

// TestAccMonitorRealBrowserResourceScreenshotDelayRemoval verifies that
// dropping screenshot_delay from the configuration converges. Uptime Kuma
// provides no way to clear the delay, so the attribute is computed and keeps
// the value the server already has instead of proposing a change forever.
func TestAccMonitorRealBrowserResourceScreenshotDelayRemoval(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRealBrowserScreenshotDelayRemoval")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRealBrowserResourceConfigScreenshotDelay(name, 3000),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("screenshot_delay"),
						knownvalue.Int64Exact(3000),
					),
				},
			},
			{
				Config: testAccMonitorRealBrowserResourceConfigWithoutScreenshotDelay(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("screenshot_delay"),
						knownvalue.Int64Exact(3000),
					),
				},
			},
			{
				// A second apply of the same configuration must still converge.
				Config: testAccMonitorRealBrowserResourceConfigWithoutScreenshotDelay(name),
			},
			{
				// The delay is disabled by setting it to 0, not by removing it.
				Config: testAccMonitorRealBrowserResourceConfigScreenshotDelay(name, 0),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_real_browser.test",
						tfjsonpath.New("screenshot_delay"),
						knownvalue.Int64Exact(0),
					),
				},
			},
		},
	})
}

// TestAccMonitorRealBrowserResourceScreenshotDelayTooLarge verifies that the
// server rejects a delay of at least half the interval.
func TestAccMonitorRealBrowserResourceScreenshotDelayTooLarge(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRealBrowserScreenshotDelayTooLarge")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// interval defaults to 60 seconds, so the limit is 30000 ms.
				Config:      testAccMonitorRealBrowserResourceConfigScreenshotDelay(name, 40000),
				ExpectError: regexp.MustCompile(`Screenshot delay must be less than`),
			},
			{
				Config:      testAccMonitorRealBrowserResourceConfigScreenshotDelay(name, -1),
				ExpectError: regexp.MustCompile(`must be at least 0`),
			},
		},
	})
}

func testAccMonitorRealBrowserResourceConfigWithoutScreenshotDelay(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_real_browser" "test" {
  name = %[1]q
  url  = "https://example.com"
}
`, name)
}
