package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMaintenanceMonitorsResource(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	monitorName1 := acctest.RandomWithPrefix("TestMonitor1")
	monitorName2 := acctest.RandomWithPrefix("TestMonitor2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceMonitorsResourceConfigSingle(maintenanceTitle, monitorName1),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_maintenance_monitors.test", tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(1)),
				},
			},
			{
				Config: testAccMaintenanceMonitorsResourceConfigMultiple(maintenanceTitle, monitorName1, monitorName2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_maintenance_monitors.test", tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(2)),
				},
			},
		},
	})
}

func testAccMaintenanceMonitorsResourceConfigSingle(maintenanceTitle, monitorName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_monitor_http" "test1" {
  name = %[2]q
  url  = "https://example.com"
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = [uptimekuma_monitor_http.test1.id]
}
`, maintenanceTitle, monitorName)
}

func testAccMaintenanceMonitorsResourceConfigMultiple(maintenanceTitle, monitorName1, monitorName2 string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_monitor_http" "test1" {
  name = %[2]q
  url  = "https://example.com"
}

resource "uptimekuma_monitor_http" "test2" {
  name = %[3]q
  url  = "https://example.org"
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = [
    uptimekuma_monitor_http.test1.id,
    uptimekuma_monitor_http.test2.id
  ]
}
`, maintenanceTitle, monitorName1, monitorName2)
}

func TestAccMaintenanceMonitorsResource_Update(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	monitorName1 := acctest.RandomWithPrefix("TestMonitor1")
	monitorName2 := acctest.RandomWithPrefix("TestMonitor2")
	monitorName3 := acctest.RandomWithPrefix("TestMonitor3")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceMonitorsResourceConfigUpdateInitial(maintenanceTitle, monitorName1, monitorName2),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_maintenance_monitors.test", tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(2)),
				},
			},
			{
				Config: testAccMaintenanceMonitorsResourceConfigUpdateChanged(maintenanceTitle, monitorName1, monitorName2, monitorName3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_maintenance_monitors.test", tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(2)),
				},
			},
		},
	})
}

func testAccMaintenanceMonitorsResourceConfigUpdateInitial(maintenanceTitle, monitorName1, monitorName2 string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_monitor_http" "test1" {
  name = %[2]q
  url  = "https://example.com"
}

resource "uptimekuma_monitor_http" "test2" {
  name = %[3]q
  url  = "https://example.org"
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = [
    uptimekuma_monitor_http.test1.id,
    uptimekuma_monitor_http.test2.id
  ]
}
`, maintenanceTitle, monitorName1, monitorName2)
}

func testAccMaintenanceMonitorsResourceConfigUpdateChanged(maintenanceTitle, monitorName1, monitorName2, monitorName3 string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_monitor_http" "test1" {
  name = %[2]q
  url  = "https://example.com"
}

resource "uptimekuma_monitor_http" "test2" {
  name = %[3]q
  url  = "https://example.org"
}

resource "uptimekuma_monitor_http" "test3" {
  name = %[4]q
  url  = "https://example.net"
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = [
    uptimekuma_monitor_http.test1.id,
    uptimekuma_monitor_http.test3.id
  ]
}
`, maintenanceTitle, monitorName1, monitorName2, monitorName3)
}

func TestAccMaintenanceMonitorsResource_Empty(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceMonitorsResourceConfigEmpty(maintenanceTitle),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_maintenance_monitors.test", tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testAccMaintenanceMonitorsResourceConfigEmpty(maintenanceTitle string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = []
}
`, maintenanceTitle)
}

func TestAccMaintenanceMonitorsResource_WithScheduledMaintenance(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestScheduledMaintenance")
	monitorName := acctest.RandomWithPrefix("TestMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceMonitorsResourceConfigWithScheduledMaintenance(maintenanceTitle, monitorName),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_maintenance.test", tfjsonpath.New("strategy"), knownvalue.StringExact("recurring-weekday")),
					statecheck.ExpectKnownValue("uptimekuma_maintenance_monitors.test", tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(1)),
				},
			},
		},
	})
}

func testAccMaintenanceMonitorsResourceConfigWithScheduledMaintenance(maintenanceTitle, monitorName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = "Weekly maintenance window"
  strategy    = "recurring-weekday"
  active      = true
  weekdays    = [1, 3, 5]
  start_time = {
    hours   = 2
    minutes = 0
    seconds = 0
  }
  end_time = {
    hours   = 4
    minutes = 0
    seconds = 0
  }
  timezone = "UTC"
}

resource "uptimekuma_monitor_http" "test" {
  name = %[2]q
  url  = "https://example.com"
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = [uptimekuma_monitor_http.test.id]
}
`, maintenanceTitle, monitorName)
}

func TestAccMaintenanceMonitorsResource_ImportBasic(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	monitorName := acctest.RandomWithPrefix("TestMonitor")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceMonitorsResourceConfigSingle(maintenanceTitle, monitorName),
			},
			{
				ResourceName:                         "uptimekuma_maintenance_monitors.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "maintenance_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["uptimekuma_maintenance_monitors.test"]
					return rs.Primary.Attributes["maintenance_id"], nil
				},
			},
		},
	})
}
