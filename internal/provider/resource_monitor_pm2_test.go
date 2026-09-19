package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccMonitorPM2Resource covers the create, import and update round-trip.
//
// The PM2 check shells out to the `pm2` CLI on the Uptime Kuma host, which the
// official container image does not ship. The monitor therefore never goes up
// against the test container, so these tests assert on the configuration
// round-trip only.
func TestAccMonitorPM2Resource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2Monitor")
	nameUpdated := acctest.RandomWithPrefix("TestPM2MonitorUpdated")
	processName := "api-server"
	processNameUpdated := "worker 1"
	description := "Test PM2 monitor with description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPM2ResourceConfigWithDescription(
					name,
					processName,
					60,
					description,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact(processName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_pm2.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// A process name containing a space is valid for PM2, unlike
				// for the system-service monitor it shares the wire field with.
				Config: testAccMonitorPM2ResourceConfigWithDescription(
					nameUpdated,
					processNameUpdated,
					120,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact(processNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("description"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorPM2ResourceConfigWithDescription(
	name string, processName string,
	interval int64, description string,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description  = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_pm2" "test" {
  name         = %[1]q
  process_name = %[2]q
%[3]s
  interval     = %[4]d
  active       = true
}
`, name, processName, descField, interval)
}

// TestAccMonitorPM2ResourceMinimal verifies the shared monitor defaults apply.
func TestAccMonitorPM2ResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2MonitorMinimal")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPM2ResourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact("api-server"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(60),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccMonitorPM2ResourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_pm2" "test" {
  name         = %[1]q
  process_name = "api-server"
}
`, name)
}

// TestAccMonitorPM2ResourceWithAllOptions verifies every supported attribute,
// including a numeric PM2 id as the process name.
func TestAccMonitorPM2ResourceWithAllOptions(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2MonitorFull")
	description := "Full PM2 monitor test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorPM2ResourceConfigWithAllOptions(name, description),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(description),
					),
					// The server matches the numeric PM2 id as a string, just
					// like the process name.
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact("0"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(90),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(0),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
			},
		},
	})
}

func testAccMonitorPM2ResourceConfigWithAllOptions(name string, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_pm2" "test" {
  name            = %[1]q
  description     = %[2]q
  process_name    = "0"
  interval        = 120
  retry_interval  = 90
  resend_interval = 0
  max_retries     = 5
  upside_down     = false
  active          = false
}
`, name, description)
}

// TestAccMonitorPM2ResourcePaddedProcessName verifies that a process name with
// surrounding whitespace applies cleanly and does not produce a perpetual diff.
// Uptime Kuma trims the value before storing it, so the provider sends the
// trimmed value and keeps the configured one in state.
func TestAccMonitorPM2ResourcePaddedProcessName(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2MonitorPadded")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorPM2ResourceConfigPadded(name),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_pm2.test",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact("  api-server  "),
					),
					// What Uptime Kuma stored, read back through the data source.
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_pm2.test",
						tfjsonpath.New("process_name"),
						knownvalue.StringExact("api-server"),
					),
				},
			},
			{
				// Re-planning the same configuration must be a no-op.
				Config:   testAccMonitorPM2ResourceConfigPadded(name),
				PlanOnly: true,
			},
		},
	})
}

func testAccMonitorPM2ResourceConfigPadded(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_pm2" "test" {
  name         = %[1]q
  process_name = "  api-server  "
}

data "uptimekuma_monitor_pm2" "test" {
  id = uptimekuma_monitor_pm2.test.id
}
`, name)
}

// TestAccMonitorPM2ResourceBlankProcessName verifies that a process name that is
// empty once trimmed is rejected at plan time, the way Uptime Kuma rejects it.
func TestAccMonitorPM2ResourceBlankProcessName(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2MonitorBlank")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorPM2ResourceConfigProcessName(name, "   "),
				ExpectError: regexp.MustCompile(`must not be empty or consist only of whitespace`),
				PlanOnly:    true,
			},
		},
	})
}

// TestAccMonitorPM2ResourceControlCharacter verifies that the ASCII control
// characters Uptime Kuma rejects are reported at plan time.
func TestAccMonitorPM2ResourceControlCharacter(t *testing.T) {
	name := acctest.RandomWithPrefix("TestPM2MonitorControlChar")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// \u0007 is BEL, which HCL turns into the control character itself.
				Config:      testAccMonitorPM2ResourceConfigProcessName(name, `api\u0007server`),
				ExpectError: regexp.MustCompile(`must not contain ASCII control characters`),
				PlanOnly:    true,
			},
		},
	})
}

func testAccMonitorPM2ResourceConfigProcessName(name string, processName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_pm2" "test" {
  name         = %[1]q
  process_name = "%[2]s"
}
`, name, processName)
}
