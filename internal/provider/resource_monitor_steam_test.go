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

func TestAccMonitorSteamResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSteamMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestSteamMonitorUpdated")
	hostname := "192.168.1.100"
	hostnameUpdated := "10.0.0.1"
	port := int64(27015)
	portUpdated := int64(27016)
	description := "Test Steam game server monitor"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSteamResourceConfigWithDescription(
					name,
					hostname,
					port,
					60,
					description,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(port),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("timeout"),
						knownvalue.Int64Exact(48),
					),
				},
			},
			{
				Config: testAccMonitorSteamResourceConfigWithDescription(
					nameUpdated,
					hostnameUpdated,
					portUpdated,
					120,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostnameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(portUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_steam.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_steam.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorSteamResourceConfigWithDescription(
	name string, hostname string,
	port int64, interval int64,
	description string,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_steam" "test" {
  name     = %[1]q
  hostname = %[2]q
  port     = %[3]d
%[4]s
  interval = %[5]d
  active   = true
}
`, name, hostname, port, descField, interval)
}
