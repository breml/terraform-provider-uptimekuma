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

func TestAccMonitorDNSDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDNSMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorDNSDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_dns.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccMonitorDNSDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_dns" "test" {
  name     = %[1]q
  hostname = "google.com"
}

data "uptimekuma_monitor_dns" "test" {
  name = uptimekuma_monitor_dns.test.name
}
`, name)
}

func TestAccMonitorDNSDataSourceByID(t *testing.T) {
	name := acctest.RandomWithPrefix("TestDNSMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorDNSDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_monitor_dns.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccMonitorDNSDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_dns" "test" {
  name     = %[1]q
  hostname = "google.com"
}

data "uptimekuma_monitor_dns" "test" {
  id = uptimekuma_monitor_dns.test.id
}
`, name)
}
