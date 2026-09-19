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

func TestAccMonitorNTPResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNTPMonitor")
	nameUpdated := acctest.RandomWithPrefix("TestNTPMonitorUpdated")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorNTPResourceConfigFull(
					name,
					"pool.ntp.org",
					123,
					300,
					"Test NTP monitor",
					5,
					1000,
					500,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("pool.ntp.org"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(123),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Test NTP monitor"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(300),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_stratum_threshold"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_time_offset_threshold"),
						knownvalue.Int64Exact(1000),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_root_dispersion_threshold"),
						knownvalue.Int64Exact(500),
					),
				},
			},
			{
				Config: testAccMonitorNTPResourceConfigFull(
					nameUpdated,
					"time.cloudflare.com",
					1123,
					600,
					"",
					3,
					250,
					100,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("time.cloudflare.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(1123),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("interval"),
						knownvalue.Int64Exact(600),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_stratum_threshold"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_time_offset_threshold"),
						knownvalue.Int64Exact(250),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_root_dispersion_threshold"),
						knownvalue.Int64Exact(100),
					),
					// Step 2 drops the description, which must read back as null.
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("description"),
						knownvalue.Null(),
					),
				},
			},
			{
				// Removing the optional fields must send an explicit null so the
				// server clears the column, rather than leaving the old value in
				// place and producing a permanent diff.
				Config:             testAccMonitorNTPResourceConfigMinimal(nameUpdated, "time.cloudflare.com"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("port"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_stratum_threshold"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_time_offset_threshold"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_root_dispersion_threshold"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_ntp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorNTPResourceConfigFull(
	name string,
	hostname string,
	port int64,
	interval int64,
	description string,
	stratum int64,
	timeOffset int64,
	rootDispersion int64,
) string {
	descField := ""
	if description != "" {
		descField = fmt.Sprintf("  description = %q", description)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ntp" "test" {
  name     = %[1]q
  hostname = %[2]q
  port     = %[3]d
%[4]s
  interval = %[5]d
  active   = true

  ntp_stratum_threshold         = %[6]d
  ntp_time_offset_threshold     = %[7]d
  ntp_root_dispersion_threshold = %[8]d
}
`, name, hostname, port, descField, interval, stratum, timeOffset, rootDispersion)
}

// TestAccMonitorNTPResourceMinimal verifies that the optional fields stay null
// when they are not configured. The server stores a nil pointer as SQL NULL and
// reads it back as null rather than as the value the check falls back to, so
// modelling the fallbacks as Terraform defaults would cause a perpetual diff.
// The timeout is the exception: the column is NOT NULL, so the client always
// sends a value and the provider defaults it to 10.
func TestAccMonitorNTPResourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNTPMonitorMinimal")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccMonitorNTPResourceConfigMinimal(name, "pool.ntp.org"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("port"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_stratum_threshold"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_time_offset_threshold"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("ntp_root_dispersion_threshold"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(10),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_ntp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorNTPResourceConfigMinimal(name string, hostname string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ntp" "test" {
  name     = %[1]q
  hostname = %[2]q
  interval = 300
  active   = true
}
`, name, hostname)
}

// TestAccMonitorNTPResourceFractionalTimeout verifies that a fractional timeout
// round-trips unchanged. Uptime Kuma stores the timeout in a floating point
// column, so NTP monitors return exactly what was configured.
func TestAccMonitorNTPResourceFractionalTimeout(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNTPMonitorFractionalTimeout")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorNTPResourceConfigWithTimeout(name, "pool.ntp.org", 1.5),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(1.5),
					),
				},
			},
			{
				Config: testAccMonitorNTPResourceConfigWithTimeout(name, "pool.ntp.org", 30.25),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(30.25),
					),
				},
			},
			{
				// Removing the attribute must return to the schema default.
				Config:             testAccMonitorNTPResourceConfigMinimal(name, "pool.ntp.org"),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(10),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_ntp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorNTPResourceConfigWithTimeout(name string, hostname string, timeout float64) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ntp" "test" {
  name     = %[1]q
  hostname = %[2]q
  timeout  = %[3]v
  interval = 300
  active   = true
}
`, name, hostname, timeout)
}

// TestAccMonitorNTPResourceWithAllOptions covers the base monitor fields, which
// every monitor type copies into its own build and populate helpers rather than
// sharing, so a slip here would silently drop a monitor out of its group or
// leave it without notifications.
func TestAccMonitorNTPResourceWithAllOptions(t *testing.T) {
	groupName := acctest.RandomWithPrefix("TestNTPMonitorGroup")
	notificationName := acctest.RandomWithPrefix("TestNTPMonitorNotification")
	tagName := acctest.RandomWithPrefix("TestNTPMonitorTag")
	name := acctest.RandomWithPrefix("TestNTPMonitorFull")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorNTPResourceConfigWithAllOptions(
					groupName, notificationName, tagName, name,
				),
				ExpectNonEmptyPlan: false,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("parent"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("notification_ids"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("tags"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("tags").AtSliceIndex(0).AtMapKey("value"),
						knownvalue.StringExact("production"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("retry_interval"),
						knownvalue.Int64Exact(120),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("resend_interval"),
						knownvalue.Int64Exact(10),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("max_retries"),
						knownvalue.Int64Exact(5),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor_ntp.test",
						tfjsonpath.New("upside_down"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor_ntp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMonitorNTPResourceConfigWithAllOptions(
	groupName string,
	notificationName string,
	tagName string,
	name string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_group" "test" {
  name = %[1]q
}

resource "uptimekuma_notification_webhook" "test" {
  name        = %[2]q
  webhook_url = "https://example.com/webhook"
  is_active   = true
}

resource "uptimekuma_tag" "test" {
  name  = %[3]q
  color = "#00ff00"
}

resource "uptimekuma_monitor_ntp" "test" {
  name             = %[4]q
  hostname         = "pool.ntp.org"
  interval         = 300
  retry_interval   = 120
  resend_interval  = 10
  max_retries      = 5
  upside_down      = true
  active           = true
  parent           = uptimekuma_monitor_group.test.id
  notification_ids = [uptimekuma_notification_webhook.test.id]

  tags = [
    {
      tag_id = uptimekuma_tag.test.id
      value  = "production"
    },
  ]
}
`, groupName, notificationName, tagName, name)
}

// TestAccMonitorNTPResourceValidators pins the plan-time bounds. The stratum
// range and the AtLeast(1) thresholds encode server behaviour rather than
// arbitrary limits: the check treats 0 as unset, so allowing it would silently
// apply the fallback instead of the configured value.
func TestAccMonitorNTPResourceValidators(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNTPMonitorValidators")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorNTPResourceConfigWithAttribute(
					name, "port", "0",
				),
				ExpectError: regexp.MustCompile(`must be between 1 and 65535`),
			},
			{
				Config: testAccMonitorNTPResourceConfigWithAttribute(
					name, "timeout", "0",
				),
				ExpectError: regexp.MustCompile(`must be between 1\.0{6}`),
			},
			{
				Config: testAccMonitorNTPResourceConfigWithAttribute(
					name, "ntp_stratum_threshold", "16",
				),
				ExpectError: regexp.MustCompile(`must be between 1 and 15`),
			},
			{
				Config: testAccMonitorNTPResourceConfigWithAttribute(
					name, "ntp_time_offset_threshold", "0",
				),
				ExpectError: regexp.MustCompile(`must be at least 1`),
			},
			{
				Config: testAccMonitorNTPResourceConfigWithAttribute(
					name, "ntp_root_dispersion_threshold", "0",
				),
				ExpectError: regexp.MustCompile(`must be at least 1`),
			},
		},
	})
}

func testAccMonitorNTPResourceConfigWithAttribute(name string, attribute string, value string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ntp" "test" {
  name     = %[1]q
  hostname = "pool.ntp.org"
  interval = 300

  %[2]s = %[3]s
}
`, name, attribute, value)
}
