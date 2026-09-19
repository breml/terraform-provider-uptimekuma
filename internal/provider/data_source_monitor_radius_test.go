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

func TestAccMonitorRadiusDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestRadiusMonitor")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorRadiusDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_name",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_name",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("testuser"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_name",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_id",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("radius.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_id",
						tfjsonpath.New("radius_username"),
						knownvalue.StringExact("testuser"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_radius.by_id",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorRadiusDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_radius" "test" {
  name                       = %[1]q
  hostname                   = "radius.example.com"
  radius_username            = "testuser"
  radius_password            = "testpass"
  radius_secret              = "testsecret"
  domain_expiry_notification = true
}

data "uptimekuma_monitor_radius" "by_name" {
  name = uptimekuma_monitor_radius.test.name
}

data "uptimekuma_monitor_radius" "by_id" {
  id = uptimekuma_monitor_radius.test.id
}
`, name)
}
