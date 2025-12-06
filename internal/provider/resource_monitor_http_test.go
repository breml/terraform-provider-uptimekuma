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

func TestAccMonitorHTTPResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestHTTPMonitorUpdated")
	url := "https://httpbin.org/status/200"
	urlUpdated := "https://httpbin.org/status/201"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorHTTPResourceConfig(name, url, "GET", 60, 48),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("method"), knownvalue.StringExact("GET")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(48)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorHTTPResourceConfig(nameUpdated, urlUpdated, "POST", 120, 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("url"), knownvalue.StringExact(urlUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("method"), knownvalue.StringExact("POST")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccMonitorHTTPResourceConfig(name, url, method string, interval, timeout int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name     = %[1]q
  url      = %[2]q
  method   = %[3]q
  interval = %[4]d
  timeout  = %[5]d
  active   = true
}
`, name, url, method, interval, timeout)
}

func TestAccMonitorHTTPResourceWithAuth(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitorWithAuth")
	url := "https://httpbin.org/basic-auth/user/pass"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPResourceConfigWithAuth(name, url, "user", "pass"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("auth_method"), knownvalue.StringExact("basic")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("basic_auth_user"), knownvalue.StringExact("user")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("basic_auth_pass"), knownvalue.StringExact("pass")),
				},
			},
		},
	})
}

func testAccMonitorHTTPResourceConfigWithAuth(name, url, user, pass string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name            = %[1]q
  url             = %[2]q
  auth_method     = "basic"
  basic_auth_user = %[3]q
  basic_auth_pass = %[4]q
}
`, name, url, user, pass)
}

func TestAccMonitorHTTPResourceWithStatusCodes(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitorWithStatusCodes")
	url := "https://httpbin.org/status/201"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPResourceConfigWithStatusCodes(name, url),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http.test", tfjsonpath.New("accepted_status_codes"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("200-299"),
							knownvalue.StringExact("301"),
						})),
				},
			},
		},
	})
}

func testAccMonitorHTTPResourceConfigWithStatusCodes(name, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name                  = %[1]q
  url                   = %[2]q
  accepted_status_codes = ["200-299", "301"]
}
`, name, url)
}

func TestAccMonitorHTTPResourceImport(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitorImport")
	url := "https://httpbin.org/status/200"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPResourceConfig(name, url, "GET", 60, 48),
			},
			{
				ResourceName:      "uptimekuma_monitor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
