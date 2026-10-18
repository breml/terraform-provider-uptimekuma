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

func TestAccMonitorSFTPDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSFTPMonitor")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSFTPDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_name",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sftp.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_name",
						tfjsonpath.New("port"),
						knownvalue.Int64Exact(2222),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_name",
						tfjsonpath.New("ssh_username"),
						knownvalue.StringExact("backup"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_id",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("sftp.example.com"),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_id",
						tfjsonpath.New("ssh_username"),
						knownvalue.StringExact("backup"),
					),
				},
			},
		},
	})
}

func testAccMonitorSFTPDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sftp" "test" {
  name         = %[1]q
  hostname     = "sftp.example.com"
  port         = 2222
  ssh_username = "backup"
  ssh_password = "s3cret"
}

data "uptimekuma_monitor_sftp" "by_name" {
  name = uptimekuma_monitor_sftp.test.name
}

data "uptimekuma_monitor_sftp" "by_id" {
  id = uptimekuma_monitor_sftp.test.id
}
`, name)
}

// TestAccMonitorSFTPDataSourceMinimal verifies the data source reports an
// unconfigured port as null instead of as the value the check falls back to,
// which is what the server stores and returns.
func TestAccMonitorSFTPDataSourceMinimal(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSFTPMonitorDSMinimal")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorSFTPDataSourceConfigMinimal(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_monitor_sftp.by_name",
						tfjsonpath.New("port"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

func testAccMonitorSFTPDataSourceConfigMinimal(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_sftp" "test" {
  name         = %[1]q
  hostname     = "sftp.example.com"
  ssh_username = "backup"
}

data "uptimekuma_monitor_sftp" "by_name" {
  name = uptimekuma_monitor_sftp.test.name
}
`, name)
}

// TestAccMonitorSFTPDataSourceWrongType verifies that looking up a monitor of
// another type by ID is reported rather than decoded into empty SFTP fields.
func TestAccMonitorSFTPDataSourceWrongType(t *testing.T) {
	name := acctest.RandomWithPrefix("TestSFTPMonitorDSWrongType")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMonitorSFTPDataSourceConfigWrongType(name),
				ExpectError: regexp.MustCompile(`Monitor type mismatch`),
			},
		},
	})
}

func testAccMonitorSFTPDataSourceConfigWrongType(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_monitor_ping" "test" {
  name     = %[1]q
  hostname = "127.0.0.1"
}

data "uptimekuma_monitor_sftp" "by_id" {
  id = uptimekuma_monitor_ping.test.id
}
`, name)
}
