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

func TestAccMonitorSNMPDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSNMPMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSNMPDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_snmp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccMonitorSNMPDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_snmp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccMonitorSNMPDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_snmp" "test" {
  name           = %[1]q
  hostname       = "192.168.1.1"
  snmp_version   = "2c"
  snmp_oid       = ".1.3.6.1.2.1.1.5.0"
  snmp_community = "public"
}

data "uptimekuma_monitor_snmp" "test" {
  name = uptimekuma_monitor_snmp.test.name
}
`, name)
}

func testAccMonitorSNMPDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_snmp" "test" {
  name           = %[1]q
  hostname       = "192.168.1.1"
  snmp_version   = "2c"
  snmp_oid       = ".1.3.6.1.2.1.1.5.0"
  snmp_community = "public"
}

data "uptimekuma_monitor_snmp" "test" {
  id = uptimekuma_monitor_snmp.test.id
}
`, name)
}
