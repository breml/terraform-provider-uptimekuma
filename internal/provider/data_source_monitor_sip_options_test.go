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

func TestAccMonitorSIPOptionsDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSIPOptionsMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSIPOptionsDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sip.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(5060),
					),
				},
			},
			{
				Config: testAccMonitorSIPOptionsDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sip.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(5060),
					),
				},
			},
		},
	})
}

func testAccMonitorSIPOptionsDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sip_options" "test" {
  name     = %[1]q
  hostname = "sip.example.com"
  port     = 5060
}

data "uptimekuma_monitor_sip_options" "test" {
  name = uptimekuma_monitor_sip_options.test.name
}
`, name)
}

func testAccMonitorSIPOptionsDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sip_options" "test" {
  name     = %[1]q
  hostname = "sip.example.com"
  port     = 5060
}

data "uptimekuma_monitor_sip_options" "test" {
  id = uptimekuma_monitor_sip_options.test.id
}
`, name)
}
