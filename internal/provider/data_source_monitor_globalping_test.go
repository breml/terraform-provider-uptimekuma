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

func TestAccMonitorGlobalpingDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestGlobalpingMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGlobalpingDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact("Europe"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("ip_family"),
						knownvalue.StringExact("ipv4"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("ping_count"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("invert_keyword"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config: testAccMonitorGlobalpingDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("location"),
						knownvalue.StringExact("Europe"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("ip_family"),
						knownvalue.StringExact("ipv4"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.test",
						tfjsonpath.New("ping_count"),
						knownvalue.Int64Exact(3),
					),
				},
			},
		},
	})
}

func testAccMonitorGlobalpingDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_globalping" "test" {
  name       = %[1]q
  subtype    = "ping"
  url        = "https://example.com"
  location   = "Europe"
  ip_family  = "ipv4"
  ping_count = 3
}

data "uptimekuma_monitor_globalping" "test" {
  name = uptimekuma_monitor_globalping.test.name
}
`, name)
}

func testAccMonitorGlobalpingDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_globalping" "test" {
  name       = %[1]q
  subtype    = "ping"
  url        = "https://example.com"
  location   = "Europe"
  ip_family  = "ipv4"
  ping_count = 3
}

data "uptimekuma_monitor_globalping" "test" {
  id = uptimekuma_monitor_globalping.test.id
}
`, name)
}
