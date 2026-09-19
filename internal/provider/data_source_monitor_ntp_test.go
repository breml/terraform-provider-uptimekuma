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

func TestAccMonitorNTPDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNTPMonitor")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorNTPDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("pool.ntp.org"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(123),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(15),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("ntp_stratum_threshold"),
						knownvalue.Int64Exact(4),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("ntp_time_offset_threshold"),
						knownvalue.Int64Exact(750),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("ntp_root_dispersion_threshold"),
						knownvalue.Int64Exact(250),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_id",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("pool.ntp.org"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_id",
						tfjsonpath.New("ntp_stratum_threshold"),
						knownvalue.Int64Exact(4),
					),
				},
			},
		},
	})
}

func testAccMonitorNTPDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ntp" "test" {
  name     = %[1]q
  hostname = "pool.ntp.org"
  port     = 123
  timeout  = 15
  interval = 300

  ntp_stratum_threshold         = 4
  ntp_time_offset_threshold     = 750
  ntp_root_dispersion_threshold = 250
}

data "uptimekuma_monitor_ntp" "by_name" {
  name = uptimekuma_monitor_ntp.test.name
}

data "uptimekuma_monitor_ntp" "by_id" {
  id = uptimekuma_monitor_ntp.test.id
}
`, name)
}

// TestAccMonitorNTPDataSourceMinimal verifies the data source reports the
// unconfigured optional fields as null instead of as the value the check falls
// back to, which is what the server stores and returns.
func TestAccMonitorNTPDataSourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNTPMonitorDSMinimal")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorNTPDataSourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("port"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("ntp_stratum_threshold"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("ntp_time_offset_threshold"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("ntp_root_dispersion_threshold"),
						knownvalue.Null(),
					),
					// The timeout column is NOT NULL, so the schema default the
					// resource applied is what comes back, not null.
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_ntp.by_name",
						tfjsonpath.New("timeout"),
						knownvalue.Float64Exact(10),
					),
				},
			},
		},
	})
}

func testAccMonitorNTPDataSourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ntp" "test" {
  name     = %[1]q
  hostname = "pool.ntp.org"
  interval = 300
}

data "uptimekuma_monitor_ntp" "by_name" {
  name = uptimekuma_monitor_ntp.test.name
}
`, name)
}

// TestAccMonitorNTPDataSourceWrongType verifies that looking up a monitor of
// another type by ID is reported rather than decoded into empty NTP fields.
func TestAccMonitorNTPDataSourceWrongType(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNTPMonitorDSWrongType")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorNTPDataSourceConfigWrongType(name),
				ExpectError: regexp.MustCompile(`Monitor type mismatch`),
			},
		},
	})
}

func testAccMonitorNTPDataSourceConfigWrongType(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = "127.0.0.1"
}

data "uptimekuma_monitor_ntp" "by_id" {
  id = uptimekuma_monitor_ping.test.id
}
`, name)
}
