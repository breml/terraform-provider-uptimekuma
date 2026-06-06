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

func TestAccMonitorGlobalpingResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGlobalpingMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestGlobalpingMonitorUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorGlobalpingResourceConfig(name, "ping", 60),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("invert_keyword"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("ip_family"),
						knownvalue.StringExact(""),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("ping_count"),
						knownvalue.Int64Exact(0),
					),
				},
			},
			{
				Config: testAccMonitorGlobalpingResourceConfig(nameUpdated, "dns", 120),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("dns"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_globalping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorGlobalpingResourceConfig(name string, subtype string, interval int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_globalping" "test" {
  name     = %[1]q
  subtype  = %[2]q
  url      = "https://example.com"
  interval = %[3]d
  active   = true
}
`, name, subtype, interval)
}

func TestAccMonitorGlobalpingResourceHTTPSubtype(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGlobalpingHTTPMonitor")
	keyword := "Example"
	keywordUpdated := "Domain"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGlobalpingHTTPSubtypeConfig(name, keyword, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("http"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("keyword"),
						knownvalue.StringExact(keyword),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("invert_keyword"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config: testAccMonitorGlobalpingHTTPSubtypeConfig(name, keywordUpdated, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("keyword"),
						knownvalue.StringExact(keywordUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("invert_keyword"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorGlobalpingHTTPSubtypeConfig(name string, keyword string, invertKeyword bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_globalping" "test" {
  name           = %[1]q
  subtype        = "http"
  url            = "https://example.com"
  keyword        = %[2]q
  invert_keyword = %[3]t
}
`, name, keyword, invertKeyword)
}

func TestAccMonitorGlobalpingResourceDNSSubtype(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGlobalpingDNSMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGlobalpingDNSSubtypeConfig(name, "example.com", "A", "1.1.1.1"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("dns"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("dns_resolve_type"),
						knownvalue.StringExact("A"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("dns_resolve_server"),
						knownvalue.StringExact("1.1.1.1"),
					),
				},
			},
			{
				Config: testAccMonitorGlobalpingDNSSubtypeConfig(name, "example.com", "AAAA", "8.8.8.8"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("dns_resolve_type"),
						knownvalue.StringExact("AAAA"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("dns_resolve_server"),
						knownvalue.StringExact("8.8.8.8"),
					),
				},
			},
		},
	})
}

func testAccMonitorGlobalpingDNSSubtypeConfig(
	name string,
	hostname string,
	dnsResolveType string,
	dnsResolveServer string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_globalping" "test" {
  name               = %[1]q
  subtype            = "dns"
  url                = "https://example.com"
  hostname           = %[2]q
  dns_resolve_type   = %[3]q
  dns_resolve_server = %[4]q
}
`, name, hostname, dnsResolveType, dnsResolveServer)
}

func TestAccMonitorGlobalpingResourceWithLocation(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGlobalpingLocationMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGlobalpingLocationConfig(name, "Europe"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact("Europe"),
					),
				},
			},
			{
				Config: testAccMonitorGlobalpingLocationConfig(name, "North America"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact("North America"),
					),
				},
			},
		},
	})
}

func testAccMonitorGlobalpingLocationConfig(name string, location string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_globalping" "test" {
  name     = %[1]q
  subtype  = "ping"
  url      = "https://example.com"
  location = %[2]q
}
`, name, location)
}

func TestAccMonitorGlobalpingResourceWithStatusCodes(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGlobalpingStatusCodesMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGlobalpingStatusCodesConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_globalping.test",
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

func testAccMonitorGlobalpingStatusCodesConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_globalping" "test" {
  name                  = %[1]q
  subtype               = "http"
  url                   = "https://example.com"
  accepted_status_codes = ["200-299", "301"]
}
`, name)
}
