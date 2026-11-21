package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/breml/go-uptime-kuma-client/maintenance"
)

func TestAccMaintenanceDataSource(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenance")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceDataSourceConfig(title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance.by_name",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance.by_id",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
				},
			},
		},
	})
}

func testAccMaintenanceDataSourceConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title       = %[1]q
  description = "Test maintenance"
  strategy    = "single"
  active      = true
  start_date  = "2025-12-31T10:00:00Z"
  end_date    = "2025-12-31T12:00:00Z"
  timezone    = "UTC"
}

data "uptimekuma_maintenance" "by_name" {
  name = uptimekuma_maintenance.test.title
}

data "uptimekuma_maintenance" "by_id" {
  id = uptimekuma_maintenance.test.id
}
`, title)
}

// TestAccMaintenanceDataSource_CreatedInUI reads a window whose dates the Uptime
// Kuma UI wrote, without UTC offset, by name, which searches the maintenanceList
// cache, and by id, which asks the server with getMaintenance (#358).
func TestAccMaintenanceDataSource_CreatedInUI(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenanceUI")
	kumaClient := testAccOutOfBandClient(t)

	var created *maintenance.Maintenance

	createdID := func(value string) error {
		want := strconv.FormatInt(created.ID, 10)
		if value != want {
			return fmt.Errorf("id = %s, want %s", value, want)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { created = testAccCreateUIMaintenance(t, kumaClient, title, "single") },
				Config:    testAccMaintenanceDataSourceConfigCreatedInUI(title),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.uptimekuma_maintenance.by_name", "id", createdID),
					resource.TestCheckResourceAttrWith("data.uptimekuma_maintenance.by_id", "id", createdID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance.by_name",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance.by_id",
						tfjsonpath.New("title"),
						knownvalue.StringExact(title),
					),
				},
			},
		},
	})
}

// testAccMaintenanceDataSourceConfigCreatedInUI looks the window up by id through
// the by-name lookup, because its id is not known before PreConfig creates it.
func testAccMaintenanceDataSourceConfigCreatedInUI(title string) string {
	return providerConfig() + fmt.Sprintf(`
data "uptimekuma_maintenance" "by_name" {
  name = %[1]q
}

data "uptimekuma_maintenance" "by_id" {
  id = data.uptimekuma_maintenance.by_name.id
}
`, title)
}

func TestAccMaintenanceDataSource_NotFoundByName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMaintenanceDataSourceConfigNotFound(),
				ExpectError: regexp.MustCompile("No maintenance window with title 'NonExistentMaintenance' found"),
			},
		},
	})
}

func testAccMaintenanceDataSourceConfigNotFound() string {
	return providerConfig() + `
data "uptimekuma_maintenance" "test" {
  name = "NonExistentMaintenance"
}
`
}

func TestAccMaintenanceDataSource_MultipleSameName(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenance")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMaintenanceDataSourceConfigMultipleSameName(title),
				ExpectError: regexp.MustCompile("Multiple maintenance windows with title"),
			},
		},
	})
}

func testAccMaintenanceDataSourceConfigMultipleSameName(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test1" {
  title       = %[1]q
  description = "First test maintenance"
  strategy    = "single"
  active      = true
  start_date  = "2025-12-31T10:00:00Z"
  end_date    = "2025-12-31T12:00:00Z"
  timezone    = "UTC"
}

resource "uptimekuma_maintenance" "test2" {
  title       = %[1]q
  description = "Second test maintenance"
  strategy    = "single"
  active      = true
  start_date  = "2025-12-31T14:00:00Z"
  end_date    = "2025-12-31T16:00:00Z"
  timezone    = "UTC"
}

data "uptimekuma_maintenance" "test" {
  name = %[1]q
  depends_on = [
    uptimekuma_maintenance.test1,
    uptimekuma_maintenance.test2,
  ]
}
`, title)
}

func TestAccMaintenanceDataSource_MissingParameters(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMaintenanceDataSourceConfigMissingParams(),
				ExpectError: regexp.MustCompile("Either 'id' or 'name' must be specified"),
			},
		},
	})
}

func testAccMaintenanceDataSourceConfigMissingParams() string {
	return providerConfig() + `
data "uptimekuma_maintenance" "test" {
}
`
}
