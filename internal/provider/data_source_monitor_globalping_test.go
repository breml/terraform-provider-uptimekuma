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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorGlobalpingDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_name",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_name",
						tfjsonpath.New("location"),
						knownvalue.StringExact("Europe"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_name",
						tfjsonpath.New("ip_family"),
						knownvalue.StringExact("ipv4"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_name",
						tfjsonpath.New("ping_count"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_name",
						tfjsonpath.New("invert_keyword"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_id",
						tfjsonpath.New("subtype"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_id",
						tfjsonpath.New("location"),
						knownvalue.StringExact("Europe"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_id",
						tfjsonpath.New("ip_family"),
						knownvalue.StringExact("ipv4"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_globalping.by_id",
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

data "uptimekuma_monitor_globalping" "by_name" {
  name = uptimekuma_monitor_globalping.test.name
}

data "uptimekuma_monitor_globalping" "by_id" {
  id = uptimekuma_monitor_globalping.test.id
}
`, name)
}
