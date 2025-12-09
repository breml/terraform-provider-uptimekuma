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

func TestAccMonitorHTTPJSONQueryResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPJSONQueryMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestHTTPJSONQueryMonitorUpdated")
	url := "https://httpbin.org/json"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorHTTPJSONQueryResourceConfig(name, url, "$.slideshow.author", "Yours Truly", "==", 60, 48),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path"), knownvalue.StringExact("$.slideshow.author")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("expected_value"), knownvalue.StringExact("Yours Truly")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path_operator"), knownvalue.StringExact("==")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(48)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorHTTPJSONQueryResourceConfig(nameUpdated, url, "$.slideshow.slides[0].title", "Wake up to WonderWidgets!", "contains", 120, 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path"), knownvalue.StringExact("$.slideshow.slides[0].title")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("expected_value"), knownvalue.StringExact("Wake up to WonderWidgets!")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path_operator"), knownvalue.StringExact("contains")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("timeout"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_http_json_query.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorHTTPJSONQueryResourceConfig(name, url, jsonPath, expectedValue, operator string, interval, timeout int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_json_query" "test" {
  name                = %[1]q
  url                 = %[2]q
  json_path           = %[3]q
  expected_value      = %[4]q
  json_path_operator  = %[5]q
  interval            = %[6]d
  timeout             = %[7]d
  active              = true
}
`, name, url, jsonPath, expectedValue, operator, interval, timeout)
}

func TestAccMonitorHTTPJSONQueryResourceWithDefaultOperator(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPJSONQueryMonitorDefault")
	url := "https://httpbin.org/json"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPJSONQueryResourceConfigWithDefaultOperator(name, url, "$.slideshow.author", "Yours Truly"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path"), knownvalue.StringExact("$.slideshow.author")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("expected_value"), knownvalue.StringExact("Yours Truly")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path_operator"), knownvalue.StringExact("==")),
				},
			},
		},
	})
}

func testAccMonitorHTTPJSONQueryResourceConfigWithDefaultOperator(name, url, jsonPath, expectedValue string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_json_query" "test" {
  name           = %[1]q
  url            = %[2]q
  json_path      = %[3]q
  expected_value = %[4]q
}
`, name, url, jsonPath, expectedValue)
}

func TestAccMonitorHTTPJSONQueryResourceWithAuth(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPJSONQueryMonitorWithAuth")
	url := "https://httpbin.org/basic-auth/user/pass"
	jsonPath := "$.authenticated"
	expectedValue := "true"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPJSONQueryResourceConfigWithAuth(name, url, jsonPath, expectedValue, "user", "pass"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path"), knownvalue.StringExact(jsonPath)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("expected_value"), knownvalue.StringExact(expectedValue)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("auth_method"), knownvalue.StringExact("basic")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("basic_auth_user"), knownvalue.StringExact("user")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("basic_auth_pass"), knownvalue.StringExact("pass")),
				},
			},
		},
	})
}

func testAccMonitorHTTPJSONQueryResourceConfigWithAuth(name, url, jsonPath, expectedValue, user, pass string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_json_query" "test" {
  name            = %[1]q
  url             = %[2]q
  json_path       = %[3]q
  expected_value  = %[4]q
  auth_method     = "basic"
  basic_auth_user = %[5]q
  basic_auth_pass = %[6]q
}
`, name, url, jsonPath, expectedValue, user, pass)
}

func TestAccMonitorHTTPJSONQueryResourceWithStatusCodes(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHTTPJSONQueryMonitorWithStatusCodes")
	url := "https://httpbin.org/json"
	jsonPath := "$.slideshow.author"
	expectedValue := "Yours Truly"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorHTTPJSONQueryResourceConfigWithStatusCodes(name, url, jsonPath, expectedValue),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("url"), knownvalue.StringExact(url)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("json_path"), knownvalue.StringExact(jsonPath)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("expected_value"), knownvalue.StringExact(expectedValue)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_http_json_query.test", tfjsonpath.New("accepted_status_codes"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("200-299"),
							knownvalue.StringExact("301"),
						})),
				},
			},
		},
	})
}

func testAccMonitorHTTPJSONQueryResourceConfigWithStatusCodes(name, url, jsonPath, expectedValue string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_http_json_query" "test" {
  name                  = %[1]q
  url                   = %[2]q
  json_path             = %[3]q
  expected_value        = %[4]q
  accepted_status_codes = ["200-299", "301"]
}
`, name, url, jsonPath, expectedValue)
}
