package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	kuma "github.com/breml/go-uptime-kuma-client"
	"github.com/breml/go-uptime-kuma-client/maintenance"
)

func TestAccMaintenanceResource_Single(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceSingle")
	titleUpdated := acctest.RandomWithPrefix("TestMaintenanceSingleUpdated")

	resource.ParallelTest(t, resource.TestCase{
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
			{
				// A window in the future has a timeslot, which the server
				// recomputes from the dates.
				Config: testAccMaintenanceResourceConfigSingleDates(
					titleUpdated, "Updated description", true, "2030-01-01T10:00:00Z", "2030-01-01T12:00:00Z",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("timeslot_list").AtSliceIndex(0).AtMapKey("end_date"),
						knownvalue.StringExact("2030-01-01T12:00:00Z"),
					),
				},
			},
			{
				// Only the end date changes, so the timeslot changes with it.
				Config: testAccMaintenanceResourceConfigSingleDates(
					titleUpdated, "Updated description", true, "2030-01-01T10:00:00Z", "2030-01-01T13:00:00Z",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("end_date"),
						knownvalue.StringExact("2030-01-01T13:00:00Z"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("timeslot_list").AtSliceIndex(0).AtMapKey("end_date"),
						knownvalue.StringExact("2030-01-01T13:00:00Z"),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigSingle(title string, description string, active bool) string {
	return testAccMaintenanceResourceConfigSingleDates(
		title, description, active, "2025-12-31T10:00:00Z", "2025-12-31T12:00:00Z",
	)
}

func testAccMaintenanceResourceConfigSingleDates(
	title string,
	description string,
	active bool,
	startDate string,
	endDate string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = %[2]q
  strategy    = "single"
  active      = %[3]t
  start_date  = %[4]q
  end_date    = %[5]q
  timezone    = "UTC"
}
`, title, description, active, startDate, endDate)
}

func TestAccMaintenanceResource_RecurringInterval(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceInterval")

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

	resource.ParallelTest(t, resource.TestCase{
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

// The windows testAccCreateUIMaintenance creates: wall-clock time in a timezone
// other than UTC, and how the provider reads those dates back.
const (
	testAccUIMaintenanceTimezone = "Europe/Zurich"
	testAccUIMaintenanceStart    = "2030-06-29T13:39:00Z"
	testAccUIMaintenanceEnd      = "2030-06-29T15:09:00Z"
	// testAccUIMaintenanceEndChanged is the end date an update moves to.
	testAccUIMaintenanceEndChanged = "2030-06-29T16:09:00Z"
	testAccUIMaintenanceCron       = "0 3 * * *"
)

// testAccUIMaintenanceFloating returns the floating date with the digits of
// 2030-06-29 at hour:minute, the day of the windows testAccCreateUIMaintenance
// creates.
func testAccUIMaintenanceFloating(hour int, minute int) maintenance.DateTime {
	return maintenance.NewFloatingDateTime(time.Date(2030, 6, 29, hour, minute, 0, 0, time.UTC))
}

// testAccCreateUIMaintenance creates a maintenance window out of band, with its
// dates the way the Uptime Kuma UI writes them: wall-clock time without UTC
// offset, such as "2030-06-29T13:39", meant in testAccUIMaintenanceTimezone.
// Before go-uptime-kuma-client decoded such dates, one window like this made the
// whole maintenanceList undecodable, so every lookup by title failed (#358).
//
// strategy is "single", with the window from testAccUIMaintenanceStart to
// testAccUIMaintenanceEnd, or "cron", running testAccUIMaintenanceCron for an
// hour within an effective date range over the same digits, which the UI
// offers for every scheduled strategy and the provider does not manage.
//
// Call it from a PreConfig, never before resource.ParallelTest: until t.Parallel
// the non-parallel tests are still running, and data_source_maintenances_test.go
// counts every window on the server.
//
// The window is deleted in t.Cleanup; deleting a window that is already gone,
// e.g. because the test imported it and Terraform destroyed it, succeeds.
func testAccCreateUIMaintenance(
	t *testing.T,
	client *kuma.Client,
	title string,
	strategy string,
) *maintenance.Maintenance {
	t.Helper()

	start := testAccUIMaintenanceFloating(13, 39)
	end := testAccUIMaintenanceFloating(15, 9)

	var m *maintenance.Maintenance

	switch strategy {
	case "single":
		m = maintenance.NewSingleMaintenance(title, "Created like the Uptime Kuma UI does",
			start.Time(), end.Time(), testAccUIMaintenanceTimezone)

	case "cron":
		m = maintenance.NewCronMaintenance(title, "Created like the Uptime Kuma UI does",
			testAccUIMaintenanceCron, 60, testAccUIMaintenanceTimezone)

	default:
		t.Fatalf("testAccCreateUIMaintenance: unsupported strategy %q", strategy)
	}

	m.DateRange = []maintenance.DateTime{start, end}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	// CreateMaintenance sets m.ID as soon as the server assigned one, also when
	// it then fails, e.g. on the fetch that follows.
	created, err := client.CreateMaintenance(ctx, m)
	if m.ID != 0 {
		id := m.ID
		t.Cleanup(func() { testAccDeleteMaintenance(t, client, id) })
	}

	// The window exists when only the update event was lost, as in
	// testAccDeleteMaintenance; the value returned then is m, so read it back.
	if errors.Is(err, kuma.ErrUpdateEventTimeout) && m.ID != 0 {
		t.Logf("maintenance window %d created, but the update event was lost: %v", m.ID, err)

		readCtx, readCancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer readCancel()

		created, err = client.GetMaintenance(readCtx, m.ID)
	}

	if err != nil {
		t.Fatalf("create maintenance window %q out of band: %v", title, err)
	}

	// Guard the premise: a server that stored these dates with an offset would
	// leave nothing of #358 for the test to exercise.
	testAccRequireFloatingRange(t, created, start, end)

	return created
}

// testAccRequireFloatingRange fails the test unless m stores start and end as
// floating dates with exactly those digits.
func testAccRequireFloatingRange(
	t *testing.T,
	m *maintenance.Maintenance,
	start maintenance.DateTime,
	end maintenance.DateTime,
) {
	t.Helper()

	err := testAccCheckFloatingRange(m, start, end)
	if err != nil {
		t.Fatal(err)
	}
}

func testAccCheckFloatingRange(m *maintenance.Maintenance, start maintenance.DateTime, end maintenance.DateTime) error {
	want := []maintenance.DateTime{start, end}

	if len(m.DateRange) != len(want) {
		return fmt.Errorf("maintenance window %d: date range %v, want floating %v", m.ID, m.DateRange, want)
	}

	for i, d := range m.DateRange {
		if !d.Floating() || !d.Equal(want[i]) {
			return fmt.Errorf("maintenance window %d: date range %v, want floating %v", m.ID, m.DateRange, want)
		}
	}

	return nil
}

// testAccCheckStoredUIRange checks out of band that the window
// testAccCreateUIMaintenance created stores start and end as floating dates.
func testAccCheckStoredUIRange(
	t *testing.T,
	client *kuma.Client,
	created **maintenance.Maintenance,
	start maintenance.DateTime,
	end maintenance.DateTime,
) resource.TestCheckFunc {
	t.Helper()

	return func(*terraform.State) error {
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
		defer cancel()

		stored, err := client.GetMaintenance(ctx, (*created).ID)
		if err != nil {
			return fmt.Errorf("read maintenance window %d out of band: %w", (*created).ID, err)
		}

		return testAccCheckFloatingRange(stored, start, end)
	}
}

// testAccDeleteMaintenance deletes a maintenance window created out of band.
// Uptime Kuma acknowledges deleting a window that is already gone, so there is
// no lookup first.
func testAccDeleteMaintenance(t *testing.T, client *kuma.Client, id int64) {
	t.Helper()

	// t.Context() is already cancelled while cleanups run.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 30*time.Second)
	defer cancel()

	err := client.DeleteMaintenance(ctx, id)
	if errors.Is(err, kuma.ErrUpdateEventTimeout) {
		t.Logf("maintenance window %d deleted, but the update event was lost: %v", id, err)
		return
	}

	if err != nil {
		t.Errorf("delete maintenance window %d: %v", id, err)
	}
}

func testAccImportUIMaintenanceID(created **maintenance.Maintenance) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return strconv.FormatInt((*created).ID, 10), nil
	}
}

// TestAccMaintenanceResource_ImportCreatedInUI imports a single window whose
// dates the Uptime Kuma UI wrote, without UTC offset, shows that a configuration
// with the same digits and a Z matches it, and that an update of another
// attribute leaves the dates floating.
func TestAccMaintenanceResource_ImportCreatedInUI(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceImportUI")
	kumaClient := testAccOutOfBandClient(t)

	var created *maintenance.Maintenance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { created = testAccCreateUIMaintenance(t, kumaClient, title, "single") },
				Config: testAccMaintenanceResourceConfigCreatedInUI(
					title, "Created like the Uptime Kuma UI does", testAccUIMaintenanceEnd,
				),
				ResourceName:       "uptimekuma_maintenance.test",
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc:  testAccImportUIMaintenanceID(&created),
				ImportStateCheck:   testAccCheckImportedUIMaintenance,
			},
			{
				Config: testAccMaintenanceResourceConfigCreatedInUI(
					title, "Created like the Uptime Kuma UI does", testAccUIMaintenanceEnd,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("start_date"),
						knownvalue.StringExact(testAccUIMaintenanceStart),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("end_date"),
						knownvalue.StringExact(testAccUIMaintenanceEnd),
					),
				},
			},
			{
				Config: testAccMaintenanceResourceConfigCreatedInUI(
					title, "Changed by Terraform", testAccUIMaintenanceEnd,
				),
				Check: testAccCheckStoredUIRange(
					t, kumaClient, &created, testAccUIMaintenanceFloating(13, 39), testAccUIMaintenanceFloating(15, 9),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Changed by Terraform"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("start_date"),
						knownvalue.StringExact(testAccUIMaintenanceStart),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("end_date"),
						knownvalue.StringExact(testAccUIMaintenanceEnd),
					),
				},
			},
			{
				// A changed date is sent floating too, with the planned digits.
				// Terraform destroys the imported window at the end of the test.
				Config: testAccMaintenanceResourceConfigCreatedInUI(
					title, "Changed by Terraform", testAccUIMaintenanceEndChanged,
				),
				Check: testAccCheckStoredUIRange(
					t, kumaClient, &created, testAccUIMaintenanceFloating(13, 39), testAccUIMaintenanceFloating(16, 9),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("start_date"),
						knownvalue.StringExact(testAccUIMaintenanceStart),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("end_date"),
						knownvalue.StringExact(testAccUIMaintenanceEndChanged),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("timeslot_list").AtSliceIndex(0).AtMapKey("end_date"),
						knownvalue.StringExact(testAccUIMaintenanceEndChanged),
					),
				},
			},
		},
	})
}

func testAccCheckImportedUIMaintenance(states []*terraform.InstanceState) error {
	if len(states) != 1 {
		return fmt.Errorf("imported %d resources, want 1", len(states))
	}

	want := map[string]string{
		"start_date":                 testAccUIMaintenanceStart,
		"end_date":                   testAccUIMaintenanceEnd,
		"timeslot_list.0.start_date": testAccUIMaintenanceStart,
		"timeslot_list.0.end_date":   testAccUIMaintenanceEnd,
		"timezone":                   testAccUIMaintenanceTimezone,
		"strategy":                   "single",
	}

	for attr, value := range want {
		got := states[0].Attributes[attr]
		if got != value {
			return fmt.Errorf("imported %s = %q, want %q", attr, got, value)
		}
	}

	return nil
}

func testAccMaintenanceResourceConfigCreatedInUI(title string, description string, endDate string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = %[2]q
  strategy    = "single"
  active      = true
  start_date  = %[3]q
  end_date    = %[4]q
  timezone    = %[5]q
}
`, title, description, testAccUIMaintenanceStart, endDate, testAccUIMaintenanceTimezone)
}

// TestAccMaintenanceResource_ImportCronCreatedInUI imports a cron window with
// an effective date range the Uptime Kuma UI wrote, which the provider does not
// manage, and shows that an update of another attribute keeps that range.
func TestAccMaintenanceResource_ImportCronCreatedInUI(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceImportCronUI")
	kumaClient := testAccOutOfBandClient(t)

	var created *maintenance.Maintenance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { created = testAccCreateUIMaintenance(t, kumaClient, title, "cron") },
				Config: testAccMaintenanceResourceConfigCronCreatedInUI(
					title,
					"Created like the Uptime Kuma UI does",
				),
				ResourceName:       "uptimekuma_maintenance.test",
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc:  testAccImportUIMaintenanceID(&created),
			},
			{
				Config: testAccMaintenanceResourceConfigCronCreatedInUI(
					title,
					"Created like the Uptime Kuma UI does",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				// Terraform destroys the imported window at the end of the test.
				Config: testAccMaintenanceResourceConfigCronCreatedInUI(title, "Changed by Terraform"),
				Check: testAccCheckStoredUIRange(
					t, kumaClient, &created, testAccUIMaintenanceFloating(13, 39), testAccUIMaintenanceFloating(15, 9),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_maintenance.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Changed by Terraform"),
					),
				},
			},
		},
	})
}

func testAccMaintenanceResourceConfigCronCreatedInUI(title string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title            = %[1]q
  description      = %[2]q
  strategy         = "cron"
  active           = true
  cron             = %[3]q
  duration_minutes = 60
  timezone         = %[4]q
}
`, title, description, testAccUIMaintenanceCron, testAccUIMaintenanceTimezone)
}
