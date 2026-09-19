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

func TestAccMonitorPingResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPingMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestPingMonitorUpdated")
	hostname := "8.8.8.8"
	hostnameUpdated := "1.1.1.1"
	description := "Test ping monitor with description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorPingResourceConfigMinimal(name, hostname, 60, 56),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(48),
					),
				},
			},
			{
				Config: testAccMonitorPingResourceConfigWithDescription(
					name,
					hostname,
					description,
					60,
					56,
					48,
					true,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("packet_size"),
						knownvalue.Int64Exact(56),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(48),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorPingResourceConfigWithDescription(
					nameUpdated, hostnameUpdated, "", 120, 64, 30, false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostnameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("packet_size"),
						knownvalue.Int64Exact(64),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(30),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_ping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorPingResourceConfigWithDescription(
	name string, hostname string, description string,
	interval int64, packetSize int64, timeout float64,
	domainExpiry bool,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name                        = %[1]q
  hostname                    = %[2]q
%[3]s
  interval                    = %[4]d
  packet_size                 = %[5]d
  timeout                     = %[6]v
  active                      = true
  domain_expiry_notification  = %[7]t
}
`, name, hostname, descField, interval, packetSize, timeout, domainExpiry)
}

func testAccMonitorPingResourceConfigMinimal(
	name string, hostname string, interval int64, packetSize int64,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name        = %[1]q
  hostname    = %[2]q
  interval    = %[3]d
  packet_size = %[4]d
  active      = true
}
`, name, hostname, interval, packetSize)
}

// TestAccMonitorPingResourceFractionalTimeout documents that ping monitors do
// not round-trip a fractional timeout. Uptime Kuma stores the timeout in a
// floating point column, but rounds the value to whole seconds before storing
// it for ping monitors, so the state Terraform reads back never matches the
// configured value.
func TestAccMonitorPingResourceFractionalTimeout(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPingMonitorFractionalTimeout")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// The server rounds 2.5 to 3 before storing it, so the plan
				// never converges and Terraform keeps proposing the change.
				Config:             testAccMonitorPingResourceConfigWithTimeout(name, "8.8.8.8", 2.5),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccMonitorPingResourceConfigWithTimeout(name, "8.8.8.8", 30),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ping.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(30),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_ping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorPingResourceConfigWithTimeout(name string, hostname string, timeout float64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = %[2]q
  timeout  = %[3]v
  interval = 60
  active   = true
}
`, name, hostname, timeout)
}
