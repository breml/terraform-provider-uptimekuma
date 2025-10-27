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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorPingResourceConfig(name, hostname, 60, 56),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("hostname"), knownvalue.StringExact(hostname)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("packet_size"), knownvalue.Int64Exact(56)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorPingResourceConfig(nameUpdated, hostnameUpdated, 120, 64),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("hostname"), knownvalue.StringExact(hostnameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("packet_size"), knownvalue.Int64Exact(64)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccMonitorPingResourceConfig(name, hostname string, interval, packetSize int64) string {
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

func TestAccMonitorPingResourceWithDescription(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPingMonitorWithDescription")
	hostname := "8.8.8.8"
	description := "Test ping monitor with description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPingResourceConfigWithDescription(name, hostname, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("hostname"), knownvalue.StringExact(hostname)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_ping.test", tfjsonpath.New("description"), knownvalue.StringExact(description)),
				},
			},
		},
	})
}

func testAccMonitorPingResourceConfigWithDescription(name, hostname, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name        = %[1]q
  hostname    = %[2]q
  description = %[3]q
}
`, name, hostname, description)
}
