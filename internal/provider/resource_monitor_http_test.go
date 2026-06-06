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
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("method"),
						knownvalue.StringExact("GET"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(48),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorHTTPResourceConfig(nameUpdated, urlUpdated, "POST", 120, 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("method"),
						knownvalue.StringExact("POST"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorHTTPResourceConfig(name string, url string, method string, interval int64, timeout int64) string {
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
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("auth_method"),
						knownvalue.StringExact("basic"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("basic_auth_user"),
						knownvalue.StringExact("user"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("basic_auth_pass"),
						knownvalue.StringExact("pass"),
					),
				},
			},
		},
	})
}

func testAccMonitorHTTPResourceConfigWithAuth(name string, url string, user string, pass string) string {
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
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
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

func testAccMonitorHTTPResourceConfigWithStatusCodes(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name                  = %[1]q
  url                   = %[2]q
  accepted_status_codes = ["200-299", "301"]
}
`, name, url)
}

func TestAccMonitorHTTPResourceWithCacheBust(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitorWithCacheBust")
	url := "https://httpbin.org/status/200"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPResourceConfigWithCacheBust(name, url, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("cache_buster"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorHTTPResourceConfigWithCacheBust(name, url, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("cache_buster"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorHTTPResourceConfigWithCacheBust(name string, url string, cacheBust bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name         = %[1]q
  url          = %[2]q
  cache_buster = %[3]t
}
`, name, url, cacheBust)
}

func TestAccMonitorHTTPResourceActiveToggle(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitorActiveToggle")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPResourceConfigActive(name, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorHTTPResourceConfigActive(name, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config: testAccMonitorHTTPResourceConfigActive(name, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			// Re-import after the toggle to confirm the server's view of `active`
			// matches Terraform state. Without the pause/resume fix this step would
			// fail because the server would still report active=true.
			{
				ResourceName:      "uptimekuma_monitor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMonitorHTTPResourceCreateInactive(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitorCreateInactive")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPResourceConfigActive(name, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
			// Import-verify confirms the server actually persisted active=false
			// (vs. the prior bug where state was set from Read of an unchanged server).
			{
				ResourceName:      "uptimekuma_monitor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorHTTPResourceConfigActive(name string, active bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name   = %[1]q
  url    = "https://httpbin.org/status/200"
  active = %[2]t
}
`, name, active)
}

func TestAccMonitorHTTPResourceWithOAuthAudience(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPMonitorOAuthAudience")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPResourceConfigWithOAuthAudience(name, "https://api.example.com"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_http.test",
						tfjsonpath.New("oauth_audience"),
						knownvalue.StringExact("https://api.example.com/resource"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_monitor_http.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_secret"},
			},
		},
	})
}

func testAccMonitorHTTPResourceConfigWithOAuthAudience(name string, url string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http" "test" {
  name                = %[1]q
  url                 = %[2]q
  auth_method         = "oauth2-cc"
  oauth_auth_method   = "client_secret_basic"
  oauth_token_url     = "https://auth.example.com/token"
  oauth_client_id     = "client-id"
  oauth_client_secret = "client-secret"
  oauth_scopes        = "read"
  oauth_audience      = "https://api.example.com/resource"
}
`, name, url)
}
