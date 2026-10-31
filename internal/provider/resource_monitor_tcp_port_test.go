package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
	description := "Test TCP port monitor with description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorTCPPortResourceConfigWithDescription(
					name,
					hostname,
					port,
					60,
					description,
					true,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(port),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMonitorTCPPortResourceConfigWithDescription(
					nameUpdated,
					hostnameUpdated,
					portUpdated,
					120,
					"",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostnameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(portUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("domain_expiry_notification"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_tcp_port.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorTCPPortResourceConfigWithDescription(
	name string, hostname string,
	port int64, interval int64,
	description string,
	domainExpiry bool,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_tcp_port" "test" {
  name                        = %[1]q
  hostname                    = %[2]q
  port                        = %[3]d
%[4]s
  interval                    = %[5]d
  active                      = true
  domain_expiry_notification  = %[6]t
}
`, name, hostname, port, descField, interval, domainExpiry)
}

func TestAccMonitorTCPPortResourceSecurityModes(t *testing.T) {
	for _, mode := range []string{"secure", "starttls", "nostarttls"} {
		t.Run(mode, func(t *testing.T) {
			name := acctest.RandomWithPrefix("TestTCPPort" + mode)
			config := testAccMonitorTCPPortResourceConfigSecurity(name, mode)

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: config,
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(
								"uptimekuma_monitor_tcp_port.test",
								tfjsonpath.New("smtp_security"),
								knownvalue.StringExact(mode),
							),
							statecheck.ExpectKnownValue(
								"uptimekuma_monitor_tcp_port.test",
								tfjsonpath.New("expiry_notification"),
								knownvalue.Bool(true),
							),
						},
					},
					{
						ResourceName:      "uptimekuma_monitor_tcp_port.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
					{
						Config: config,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
						},
					},
				},
			})
		})
	}
}

// TestAccMonitorTCPPortResourceSecurityUnset walks the attribute from unset to
// set and back. The test framework plans again after each apply with a refresh,
// so a server that kept the previous value would fail the step with a diff.
func TestAccMonitorTCPPortResourceSecurityUnset(t *testing.T) {
	name := acctest.RandomWithPrefix("TestTCPPortSecurityUnset")
	unset := providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_tcp_port" "test" {
  name     = %[1]q
  hostname = "example.com"
  port     = 443
}
`, name)
	nullChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(
			"uptimekuma_monitor_tcp_port.test",
			tfjsonpath.New("smtp_security"),
			knownvalue.Null(),
		),
		statecheck.ExpectKnownValue(
			"uptimekuma_monitor_tcp_port.test",
			tfjsonpath.New("expiry_notification"),
			knownvalue.Bool(false),
		),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: unset, ConfigStateChecks: nullChecks},
			{
				Config: testAccMonitorTCPPortResourceConfigSecurity(name, "secure"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_tcp_port.test",
						tfjsonpath.New("smtp_security"),
						knownvalue.StringExact("secure"),
					),
				},
			},
			{Config: unset, ConfigStateChecks: nullChecks},
		},
	})
}

func TestAccMonitorTCPPortResourceInvalidSecurity(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorTCPPortResourceConfigSecurity("unused", "TLS"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

func testAccMonitorTCPPortResourceConfigSecurity(name string, security string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_tcp_port" "test" {
  name                = %[1]q
  hostname            = "example.com"
  port                = 443
  smtp_security       = %[2]q
  expiry_notification = true
}
`, name, security)
}
