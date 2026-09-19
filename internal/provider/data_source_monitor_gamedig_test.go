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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGameDigDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.by_name",
						tfjsonpath.New("game"),
						knownvalue.StringExact("minecraft"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.by_name",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.by_id",
						tfjsonpath.New("game"),
						knownvalue.StringExact("minecraft"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_gamedig.by_id",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorGameDigDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_gamedig" "test" {
  name                       = %[1]q
  hostname                   = "192.168.1.100"
  port                       = 25565
  game                       = "minecraft"
  domain_expiry_notification = true
}

data "uptimekuma_monitor_gamedig" "by_name" {
  name = uptimekuma_monitor_gamedig.test.name
}

data "uptimekuma_monitor_gamedig" "by_id" {
  id = uptimekuma_monitor_gamedig.test.id
}
`, name)
}
