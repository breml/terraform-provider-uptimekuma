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

func TestAccMonitorSIPOptionsResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSIPOptionsMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestSIPOptionsMonitorUpdated")
	hostname := "sipserver.example.com"
	hostnameUpdated := "sipserver2.example.com"
	port := int64(5060)
	portUpdated := int64(5061)
	description := "Test SIP Options monitor with description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSIPOptionsResourceConfigWithDescription(
					name,
					hostname,
					port,
					60,
					description,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(port),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_sip_options.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorSIPOptionsResourceConfigWithDescription(
					nameUpdated,
					hostnameUpdated,
					portUpdated,
					120,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostnameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(portUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("description"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorSIPOptionsResourceConfigWithDescription(
	name string, hostname string,
	port int64, interval int64,
	description string,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sip_options" "test" {
  name     = %[1]q
  hostname = %[2]q
  port     = %[3]d
%[4]s
  interval = %[5]d
  active   = true
}
`, name, hostname, port, descField, interval)
}

func TestAccMonitorSIPOptionsResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSIPOptionsMonitorMinimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSIPOptionsResourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sip.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(5060),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorSIPOptionsResourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sip_options" "test" {
  name     = %[1]q
  hostname = "sip.example.com"
  port     = 5060
}
`, name)
}

func TestAccMonitorSIPOptionsResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSIPOptionsMonitorFull")
	description := "Full SIP Options monitor test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSIPOptionsResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sip.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(5060),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(90),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_sip_options.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorSIPOptionsResourceConfigWithAllOptions(name string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sip_options" "test" {
  name            = %[1]q
  description     = %[2]q
  hostname        = "sip.example.com"
  port            = 5060
  interval        = 120
  retry_interval  = 90
  resend_interval = 0
  max_retries     = 5
  upside_down     = false
  active          = false
}
`, name, description)
}
