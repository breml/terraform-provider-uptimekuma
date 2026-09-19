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
