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

func TestAccMonitorGameDigDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGameDigMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGameDigDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("game"),
						knownvalue.StringExact("minecraft"),
					),
				},
			},
			{
				Config: testAccMonitorGameDigDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.test",
						tfjsonpath.New("game"),
						knownvalue.StringExact("minecraft"),
					),
				},
			},
		},
	})
}

func testAccMonitorGameDigDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_gamedig" "test" {
  name     = %[1]q
  hostname = "192.168.1.100"
  port     = 25565
  game     = "minecraft"
}

data "uptimekuma_monitor_gamedig" "test" {
  name = uptimekuma_monitor_gamedig.test.name
}
`, name)
}

func testAccMonitorGameDigDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_gamedig" "test" {
  name     = %[1]q
  hostname = "192.168.1.100"
  port     = 25565
  game     = "minecraft"
}

data "uptimekuma_monitor_gamedig" "test" {
  id = uptimekuma_monitor_gamedig.test.id
}
`, name)
}
