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

func TestAccMonitorTCPPortResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestTCPPortMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestTCPPortMonitorUpdated")
	hostname := "8.8.8.8"
	hostnameUpdated := "1.1.1.1"
	port := int64(443)
	portUpdated := int64(80)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorTCPPortResourceConfig(name, hostname, port, 60),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("hostname"), knownvalue.StringExact(hostname)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("port"), knownvalue.Int64Exact(port)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(60)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccMonitorTCPPortResourceConfig(nameUpdated, hostnameUpdated, portUpdated, 120),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("hostname"), knownvalue.StringExact(hostnameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("port"), knownvalue.Int64Exact(portUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("interval"), knownvalue.Int64Exact(120)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func testAccMonitorTCPPortResourceConfig(name, hostname string, port, interval int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_tcp_port" "test" {
  name     = %[1]q
  hostname = %[2]q
  port     = %[3]d
  interval = %[4]d
  active   = true
}
`, name, hostname, port, interval)
}

func TestAccMonitorTCPPortResourceWithDescription(t *testing.T) {
	name := acctest.RandomWithPrefix("TestTCPPortMonitorWithDescription")
	hostname := "example.com"
	port := int64(8080)
	description := "Test TCP port monitor with description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorTCPPortResourceConfigWithDescription(name, hostname, port, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("hostname"), knownvalue.StringExact(hostname)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("port"), knownvalue.Int64Exact(port)),
					statecheck.ExpectKnownValue("uptimekuma_monitor_tcp_port.test", tfjsonpath.New("description"), knownvalue.StringExact(description)),
				},
			},
		},
	})
}

func testAccMonitorTCPPortResourceConfigWithDescription(name, hostname string, port int64, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_tcp_port" "test" {
  name        = %[1]q
  hostname    = %[2]q
  port        = %[3]d
  description = %[4]q
}
`, name, hostname, port, description)
}
