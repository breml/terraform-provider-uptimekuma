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

func TestAccMonitorSystemServiceResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSystemServiceMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestSystemServiceMonitorUpdated")
	serviceName := "nginx.service"
	serviceNameUpdated := "sshd.service"
	description := "Test System Service monitor with description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSystemServiceResourceConfigWithDescription(
					name,
					serviceName,
					60,
					description,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("system_service_name"),
						knownvalue.StringExact(serviceName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_system_service.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMonitorSystemServiceResourceConfigWithDescription(
					nameUpdated,
					serviceNameUpdated,
					120,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("system_service_name"),
						knownvalue.StringExact(serviceNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("description"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorSystemServiceResourceConfigWithDescription(
	name string, serviceName string,
	interval int64, description string,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_system_service" "test" {
  name                = %[1]q
  system_service_name = %[2]q
%[3]s
  interval            = %[4]d
  active              = true
}
`, name, serviceName, descField, interval)
}

func TestAccMonitorSystemServiceResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSystemServiceMonitorMinimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSystemServiceResourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("system_service_name"),
						knownvalue.StringExact("nginx.service"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorSystemServiceResourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_system_service" "test" {
  name                = %[1]q
  system_service_name = "nginx.service"
}
`, name)
}

func TestAccMonitorSystemServiceResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSystemServiceMonitorFull")
	description := "Full System Service monitor test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSystemServiceResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("system_service_name"),
						knownvalue.StringExact("sshd@0.service"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(90),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_system_service.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorSystemServiceResourceConfigWithAllOptions(name string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_system_service" "test" {
  name                = %[1]q
  description         = %[2]q
  system_service_name = "sshd@0.service"
  interval            = 120
  retry_interval      = 90
  resend_interval     = 0
  max_retries         = 5
  upside_down         = false
  active              = false
}
`, name, description)
}
