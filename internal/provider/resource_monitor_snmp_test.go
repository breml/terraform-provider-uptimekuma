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

func TestAccMonitorSNMPResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSNMPMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestSNMPMonitorUpdated")
	description := "Test SNMP monitor description"
	descriptionUpdated := "Updated test SNMP monitor description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSNMPResourceConfig(
					name,
					description,
					"192.168.1.1",
					"2c",
					".1.3.6.1.2.1.1.5.0",
					"public",
					161,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("192.168.1.1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_version"),
						knownvalue.StringExact("2c"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_oid"),
						knownvalue.StringExact(".1.3.6.1.2.1.1.5.0"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(161),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorSNMPResourceConfig(
					nameUpdated,
					descriptionUpdated,
					"192.168.1.2",
					"3",
					".1.3.6.1.2.1.1.3.0",
					"private",
					161,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(descriptionUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("192.168.1.2"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_version"),
						knownvalue.StringExact("3"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_oid"),
						knownvalue.StringExact(".1.3.6.1.2.1.1.3.0"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_snmp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorSNMPResourceConfig(
	name string,
	description string,
	hostname string,
	snmpVersion string,
	snmpOID string,
	snmpCommunity string,
	port int64,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_snmp" "test" {
  name            = %[1]q
  description     = %[2]q
  hostname        = %[3]q
  snmp_version    = %[4]q
  snmp_oid        = %[5]q
  snmp_community  = %[6]q
  port            = %[7]d
  active          = true
}
`, name, description, hostname, snmpVersion, snmpOID, snmpCommunity, port)
}

func TestAccMonitorSNMPResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSNMPMonitorMinimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSNMPResourceConfigMinimal(
					name,
					"192.168.1.1",
					"2c",
					".1.3.6.1.2.1.1.5.0",
					"public",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("192.168.1.1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_version"),
						knownvalue.StringExact("2c"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_oid"),
						knownvalue.StringExact(".1.3.6.1.2.1.1.5.0"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(161),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorSNMPResourceConfigMinimal(
	name string,
	hostname string,
	snmpVersion string,
	snmpOID string,
	snmpCommunity string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_snmp" "test" {
  name           = %[1]q
  hostname       = %[2]q
  snmp_version   = %[3]q
  snmp_oid       = %[4]q
  snmp_community = %[5]q
}
`, name, hostname, snmpVersion, snmpOID, snmpCommunity)
}

func TestAccMonitorSNMPResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSNMPMonitorFull")
	description := "Full test SNMP monitor"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSNMPResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("192.168.1.1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_version"),
						knownvalue.StringExact("2c"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("snmp_oid"),
						knownvalue.StringExact(".1.3.6.1.2.1.1.5.0"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(161),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_snmp.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorSNMPResourceConfigWithAllOptions(name string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_snmp" "test" {
  name            = %[1]q
  description     = %[2]q
  hostname        = "192.168.1.1"
  snmp_version    = "2c"
  snmp_oid        = ".1.3.6.1.2.1.1.5.0"
  snmp_community  = "public"
  port            = 161
  interval        = 120
  retry_interval  = 60
  max_retries     = 5
  active          = true
  upside_down     = false
}
`, name, description)
}
