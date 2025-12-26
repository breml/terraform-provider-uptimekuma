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

func TestAccMaintenanceStatusPagesResource(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	statusPageSlug1 := acctest.RandomWithPrefix("test-status-1")
	statusPageSlug2 := acctest.RandomWithPrefix("test-status-2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceStatusPagesResourceConfigSingle(
					maintenanceTitle,
					statusPageSlug1,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
			{
				Config: testAccMaintenanceStatusPagesResourceConfigMultiple(
					maintenanceTitle,
					statusPageSlug1,
					statusPageSlug2,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
		},
	})
}

func testAccMaintenanceStatusPagesResourceConfigSingle(maintenanceTitle, statusPageSlug string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_status_page" "test1" {
  slug  = %[2]q
  title = "Test Status Page 1"
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = [uptimekuma_status_page.test1.id]
}
`, maintenanceTitle, statusPageSlug)
}

func testAccMaintenanceStatusPagesResourceConfigMultiple(
	maintenanceTitle, statusPageSlug1, statusPageSlug2 string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_status_page" "test1" {
  slug  = %[2]q
  title = "Test Status Page 1"
}

resource "uptimekuma_status_page" "test2" {
  slug  = %[3]q
  title = "Test Status Page 2"
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = [
    uptimekuma_status_page.test1.id,
    uptimekuma_status_page.test2.id
  ]
}
`, maintenanceTitle, statusPageSlug1, statusPageSlug2)
}

func TestAccMaintenanceStatusPagesResource_Import(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	statusPageSlug := acctest.RandomWithPrefix("test-status")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceStatusPagesResourceConfigSingle(maintenanceTitle, statusPageSlug),
			},
			{
				ResourceName:                         "uptimekuma_maintenance_status_pages.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "maintenance_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["uptimekuma_maintenance_status_pages.test"]
					return rs.Primary.Attributes["maintenance_id"], nil
				},
			},
		},
	})
}

func TestAccMaintenanceStatusPagesResource_Update(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	statusPageSlug1 := acctest.RandomWithPrefix("test-status-1")
	statusPageSlug2 := acctest.RandomWithPrefix("test-status-2")
	statusPageSlug3 := acctest.RandomWithPrefix("test-status-3")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceStatusPagesResourceConfigUpdateInitial(
					maintenanceTitle,
					statusPageSlug1,
					statusPageSlug2,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
			{
				Config: testAccMaintenanceStatusPagesResourceConfigUpdateChanged(
					maintenanceTitle,
					statusPageSlug1,
					statusPageSlug2,
					statusPageSlug3,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
		},
	})
}

func testAccMaintenanceStatusPagesResourceConfigUpdateInitial(
	maintenanceTitle, statusPageSlug1, statusPageSlug2 string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_status_page" "test1" {
  slug  = %[2]q
  title = "Test Status Page 1"
}

resource "uptimekuma_status_page" "test2" {
  slug  = %[3]q
  title = "Test Status Page 2"
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = [
    uptimekuma_status_page.test1.id,
    uptimekuma_status_page.test2.id
  ]
}
`, maintenanceTitle, statusPageSlug1, statusPageSlug2)
}

func testAccMaintenanceStatusPagesResourceConfigUpdateChanged(
	maintenanceTitle, statusPageSlug1, statusPageSlug2, statusPageSlug3 string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_status_page" "test1" {
  slug  = %[2]q
  title = "Test Status Page 1"
}

resource "uptimekuma_status_page" "test2" {
  slug  = %[3]q
  title = "Test Status Page 2"
}

resource "uptimekuma_status_page" "test3" {
  slug  = %[4]q
  title = "Test Status Page 3"
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = [
    uptimekuma_status_page.test1.id,
    uptimekuma_status_page.test3.id
  ]
}
`, maintenanceTitle, statusPageSlug1, statusPageSlug2, statusPageSlug3)
}

func TestAccMaintenanceStatusPagesResource_Empty(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceStatusPagesResourceConfigEmpty(maintenanceTitle),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

func testAccMaintenanceStatusPagesResourceConfigEmpty(maintenanceTitle string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = []
}
`, maintenanceTitle)
}

func TestAccMaintenanceStatusPagesResource_WithScheduledMaintenance(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestScheduledMaintenance")
	statusPageSlug := acctest.RandomWithPrefix("test-status")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceStatusPagesResourceConfigWithScheduledMaintenance(
					maintenanceTitle,
					statusPageSlug,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("recurring-interval"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
		},
	})
}

func testAccMaintenanceStatusPagesResourceConfigWithScheduledMaintenance(
	maintenanceTitle, statusPageSlug string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title        = %[1]q
  description  = "Regular maintenance window"
  strategy     = "recurring-interval"
  active       = true
  interval_day = 7
  start_time = {
    hours   = 3
    minutes = 0
    seconds = 0
  }
  end_time = {
    hours   = 5
    minutes = 0
    seconds = 0
  }
  timezone = "UTC"
}

resource "uptimekuma_status_page" "test" {
  slug      = %[2]q
  title     = "Test Status Page"
  published = true
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = [uptimekuma_status_page.test.id]
}
`, maintenanceTitle, statusPageSlug)
}

func TestAccMaintenanceStatusPagesResource_Combined(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	monitorName := acctest.RandomWithPrefix("TestMonitor")
	statusPageSlug := acctest.RandomWithPrefix("test-status")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceStatusPagesResourceConfigCombined(
					maintenanceTitle,
					monitorName,
					statusPageSlug,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_maintenance_monitors.test", tfjsonpath.New("monitor_ids"),
						knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
		},
	})
}

func testAccMaintenanceStatusPagesResourceConfigCombined(maintenanceTitle, monitorName, statusPageSlug string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = "Combined maintenance window"
  strategy    = "single"
  active      = true
  start_date  = "2025-12-31T10:00:00Z"
  end_date    = "2025-12-31T12:00:00Z"
  timezone    = "UTC"
}

resource "uptimekuma_monitor_http" "test" {
  name = %[2]q
  url  = "https://example.com"
}

resource "uptimekuma_status_page" "test" {
  slug      = %[3]q
  title     = "Test Status Page"
  published = true
}

resource "uptimekuma_maintenance_monitors" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  monitor_ids    = [uptimekuma_monitor_http.test.id]
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = [uptimekuma_status_page.test.id]
}
`, maintenanceTitle, monitorName, statusPageSlug)
}
