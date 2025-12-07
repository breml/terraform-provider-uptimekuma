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

func TestAccMonitorHTTPKeywordResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPKeywordMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestHTTPKeywordMonitorUpdated")
	url := "https://httpbin.org/html"
	keyword := "Herman"
	keywordUpdated := "Moby"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorHTTPKeywordResourceConfig(name, url, keyword, false, 60, 48),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keyword)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("invert_keyword"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("method"), knownvalue.StringExact("GET")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(48)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorHTTPKeywordResourceConfig(nameUpdated, url, keywordUpdated, false, 120, 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keywordUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("invert_keyword"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_http_keyword.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorHTTPKeywordResourceConfig(name, url, keyword string, invertKeyword bool, interval, timeout int64) string { //nolint:unparam
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_keyword" "test" {
  name           = %[1]q
  url            = %[2]q
  keyword        = %[3]q
  invert_keyword = %[4]t
  interval       = %[5]d
  timeout        = %[6]d
  active         = true
}
`, name, url, keyword, invertKeyword, interval, timeout)
}

func TestAccMonitorHTTPKeywordResourceWithInvert(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPKeywordMonitorInvert")
	url := "https://httpbin.org/html"
	keyword := "NonExistentKeyword"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPKeywordResourceConfig(name, url, keyword, true, 60, 48),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keyword)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("invert_keyword"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccMonitorHTTPKeywordResourceWithAuth(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPKeywordMonitorWithAuth")
	url := "https://httpbin.org/basic-auth/user/pass"
	keyword := "authenticated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPKeywordResourceConfigWithAuth(name, url, keyword, "user", "pass"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keyword)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("auth_method"), knownvalue.StringExact("basic")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("basic_auth_user"), knownvalue.StringExact("user")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("basic_auth_pass"), knownvalue.StringExact("pass")),
				},
			},
		},
	})
}

func testAccMonitorHTTPKeywordResourceConfigWithAuth(name, url, keyword, user, pass string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_keyword" "test" {
  name            = %[1]q
  url             = %[2]q
  keyword         = %[3]q
  auth_method     = "basic"
  basic_auth_user = %[4]q
  basic_auth_pass = %[5]q
}
`, name, url, keyword, user, pass)
}

func TestAccMonitorHTTPKeywordResourceWithStatusCodes(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPKeywordMonitorWithStatusCodes")
	url := "https://httpbin.org/html"
	keyword := "Herman"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPKeywordResourceConfigWithStatusCodes(name, url, keyword),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("keyword"), knownvalue.StringExact(keyword)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_keyword.test", tfjsonpath.New("accepted_status_codes"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("200-299"),
							knownvalue.StringExact("301"),
						})),
				},
			},
		},
	})
}

func testAccMonitorHTTPKeywordResourceConfigWithStatusCodes(name, url, keyword string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_keyword" "test" {
  name                  = %[1]q
  url                   = %[2]q
  keyword               = %[3]q
  accepted_status_codes = ["200-299", "301"]
}
`, name, url, keyword)
}
