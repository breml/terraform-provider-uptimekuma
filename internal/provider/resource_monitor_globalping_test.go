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
				Config: testAccMonitorGlobalpingResourceConfig(name, "ping", 60),
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
