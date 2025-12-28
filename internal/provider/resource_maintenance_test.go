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

func TestAccMaintenanceResource_Single(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceSingle")
	titleUpdated := acctest.RandomWithPrefix("TestMaintenanceSingleUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigSingle(title, "Test single maintenance", true),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Test single maintenance"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("single"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("timezone"),
						knownvalue.StringExact("UTC"),
					),
				},
			},
			{
				Config: testAccMaintenanceResourceConfigSingle(titleUpdated, "Updated description", false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(titleUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated description"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigSingle(title string, description string, active bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = %[2]q
  strategy    = "single"
  active      = %[3]t
  start_date  = "2025-12-31T10:00:00Z"
  end_date    = "2025-12-31T12:00:00Z"
  timezone    = "UTC"
}
`, title, description, active)
}

func TestAccMaintenanceResource_RecurringInterval(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceInterval")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigRecurringInterval(title, 7),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("recurring-interval"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("interval_day"),
						knownvalue.Int64Exact(7),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMaintenanceResourceConfigRecurringInterval(title, 14),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("interval_day"),
						knownvalue.Int64Exact(14),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigRecurringInterval(title string, intervalDay int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title        = %[1]q
  description  = "Recurring every N days"
  strategy     = "recurring-interval"
  active       = true
  interval_day = %[2]d
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
`, title, intervalDay)
}

func TestAccMaintenanceResource_RecurringWeekday(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceWeekday")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigRecurringWeekday(title),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("recurring-weekday"),
					),
					statecheck.ExpectKnownValue("uptimekuma_maintenance.test", tfjsonpath.New("weekdays"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.Int64Exact(1),
							knownvalue.Int64Exact(3),
							knownvalue.Int64Exact(5),
						})),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigRecurringWeekday(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = "Recurring on specific weekdays"
  strategy    = "recurring-weekday"
  active      = true
  weekdays    = [1, 3, 5]
  start_time = {
    hours   = 22
    minutes = 0
    seconds = 0
  }
  end_time = {
    hours   = 6
    minutes = 0
    seconds = 0
  }
  timezone = "UTC"
}
`, title)
}

func TestAccMaintenanceResource_RecurringDayOfMonth(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceDayOfMonth")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigRecurringDayOfMonth(title),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("recurring-day-of-month"),
					),
					statecheck.ExpectKnownValue("uptimekuma_maintenance.test", tfjsonpath.New("days_of_month"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("1"),
							knownvalue.StringExact("15"),
							knownvalue.StringExact("lastDay1"),
						})),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigRecurringDayOfMonth(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title         = %[1]q
  description   = "Recurring on specific days of month"
  strategy      = "recurring-day-of-month"
  active        = true
  days_of_month = ["1", "15", "lastDay1"]
  start_time = {
    hours   = 1
    minutes = 0
    seconds = 0
  }
  end_time = {
    hours   = 3
    minutes = 0
    seconds = 0
  }
  timezone = "UTC"
}
`, title)
}

func TestAccMaintenanceResource_Cron(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceCron")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigCron(title, "0 2 * * *", 120),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("cron"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("cron"),
						knownvalue.StringExact("0 2 * * *"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("duration_minutes"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccMaintenanceResourceConfigCron(title, "0 3 * * *", 60),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("cron"),
						knownvalue.StringExact("0 3 * * *"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("duration_minutes"),
						knownvalue.Int64Exact(60),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigCron(title string, cronExpr string, durationMinutes int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title            = %[1]q
  description      = "Cron-based maintenance"
  strategy         = "cron"
  active           = true
  cron             = %[2]q
  duration_minutes = %[3]d
  timezone         = "UTC"
}
`, title, cronExpr, durationMinutes)
}

func TestAccMaintenanceResource_Manual(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceManual")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigManual(title),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("manual"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigManual(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = "Manual trigger maintenance"
  strategy    = "manual"
  active      = true
  timezone    = "UTC"
}
`, title)
}

func TestAccMaintenanceResource_WithTimezone(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceTimezone")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigWithTimezone(title, "America/New_York"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("timezone"),
						knownvalue.StringExact("America/New_York"),
					),
				},
			},
			{
				Config: testAccMaintenanceResourceConfigWithTimezone(title, "Europe/London"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("timezone"),
						knownvalue.StringExact("Europe/London"),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigWithTimezone(title string, timezone string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = "Maintenance with custom timezone"
  strategy    = "single"
  active      = true
  start_date  = "2025-12-31T10:00:00Z"
  end_date    = "2025-12-31T12:00:00Z"
  timezone    = %[2]q
}
`, title, timezone)
}

func TestAccMaintenanceResource_Minimal(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceMinimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMaintenanceResourceConfigMinimal(title),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("strategy"),
						knownvalue.StringExact("manual"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("timezone"),
						knownvalue.StringExact("UTC"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(""),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigMinimal(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}
`, title)
}
