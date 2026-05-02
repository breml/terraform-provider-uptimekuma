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

func TestAccMonitorGameDigResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGameDigMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestGameDigMonitorUpdated")
	hostname := "192.168.1.100"
	hostnameUpdated := "10.0.0.1"
	port := int64(25565)
	portUpdated := int64(27015)
	game := "minecraft"
	gameUpdated := "csgo"
	description := "Test GameDig game server monitor"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGameDigResourceConfigWithDescription(
					name,
					hostname,
					port,
					game,
					60,
					description,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(port),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("game"),
						knownvalue.StringExact(game),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("gamedig_given_port_only"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorGameDigResourceConfigWithDescription(
					nameUpdated,
					hostnameUpdated,
					portUpdated,
					gameUpdated,
					120,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostnameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(portUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("game"),
						knownvalue.StringExact(gameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_gamedig.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorGameDigResourceConfigWithDescription(
	name string, hostname string,
	port int64, game string,
	interval int64, description string,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_gamedig" "test" {
  name     = %[1]q
  hostname = %[2]q
  port     = %[3]d
  game     = %[4]q
%[5]s
  interval = %[6]d
  active   = true
}
`, name, hostname, port, game, descField, interval)
}
