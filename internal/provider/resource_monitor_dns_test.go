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

func TestAccMonitorDNSResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDNSMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestDNSMonitorUpdated")
	description := "Test DNS monitor description"
	descriptionUpdated := "Updated test DNS monitor description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorDNSResourceConfig(name, description, "example.com", "A", "1.1.1.1", 53),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("hostname"), knownvalue.StringExact("example.com")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_server"), knownvalue.StringExact("1.1.1.1")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("port"), knownvalue.Int64Exact(53)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorDNSResourceConfig(nameUpdated, descriptionUpdated, "google.com", "AAAA", "8.8.8.8", 53),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("description"), knownvalue.StringExact(descriptionUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("hostname"), knownvalue.StringExact("google.com")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_type"), knownvalue.StringExact("AAAA")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_server"), knownvalue.StringExact("8.8.8.8")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccMonitorDNSResourceConfig(name, description, hostname, resolveType, server string, port int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_dns" "test" {
  name               = %[1]q
  description        = %[2]q
  hostname           = %[3]q
  dns_resolve_type   = %[4]q
  dns_resolve_server = %[5]q
  port               = %[6]d
  active             = true
}
`, name, description, hostname, resolveType, server, port)
}

func TestAccMonitorDNSResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDNSMonitorMinimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorDNSResourceConfigMinimal(name, "example.com"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("hostname"), knownvalue.StringExact("example.com")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_server"), knownvalue.StringExact("1.1.1.1")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("port"), knownvalue.Int64Exact(53)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("retry_interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("max_retries"), knownvalue.Int64Exact(3)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccMonitorDNSResourceConfigMinimal(name, hostname string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_dns" "test" {
  name     = %[1]q
  hostname = %[2]q
}
`, name, hostname)
}

func TestAccMonitorDNSResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDNSMonitorFull")
	description := "Full test DNS monitor"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorDNSResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("description"), knownvalue.StringExact(description)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("hostname"), knownvalue.StringExact("example.com")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_type"), knownvalue.StringExact("MX")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_server"), knownvalue.StringExact("8.8.8.8")),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("port"), knownvalue.Int64Exact(53)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("retry_interval"), knownvalue.Int64Exact(90)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("resend_interval"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("max_retries"), knownvalue.Int64Exact(5)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("upside_down"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("active"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccMonitorDNSResourceConfigWithAllOptions(name, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_dns" "test" {
  name               = %[1]q
  description        = %[2]q
  hostname           = "example.com"
  dns_resolve_type   = "MX"
  dns_resolve_server = "8.8.8.8"
  port               = 53
  interval           = 120
  retry_interval     = 90
  resend_interval    = 0
  max_retries        = 5
  upside_down        = false
  active             = false
}
`, name, description)
}

func TestAccMonitorDNSResourceDifferentRecordTypes(t *testing.T) {
	recordTypes := []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "SOA", "SRV", "PTR", "CAA"}

	for _, recordType := range recordTypes {
		t.Run(recordType, func(t *testing.T) {
			name := acctest.RandomWithPrefix(fmt.Sprintf("TestDNS%s", recordType))

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccMonitorDNSResourceConfigRecordType(name, "example.com", recordType),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
							statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("hostname"), knownvalue.StringExact("example.com")),
							statecheck.ExpectKnownValue("uptimekuma_monitor_dns.test", tfjsonpath.New("dns_resolve_type"), knownvalue.StringExact(recordType)),
						},
					},
				},
			})
		})
	}
}

func testAccMonitorDNSResourceConfigRecordType(name, hostname, recordType string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_dns" "test" {
  name             = %[1]q
  hostname         = %[2]q
  dns_resolve_type = %[3]q
}
`, name, hostname, recordType)
}
