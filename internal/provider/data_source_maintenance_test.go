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

func TestAccMaintenanceDataSource(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceDataSourceConfig(title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_maintenance.test", tfjsonpath.New("name"), knownvalue.StringExact(title)),
					statecheck.ExpectKnownValue("data.uptimekuma_maintenance.test", tfjsonpath.New("title"), knownvalue.StringExact(title)),
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

data "uptimekuma_maintenance" "test" {
  name = uptimekuma_maintenance.test.title
}
`, title)
}

func TestAccMaintenanceDataSourceByID(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceDataSourceConfigByID(title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_maintenance.test", tfjsonpath.New("name"), knownvalue.StringExact(title)),
					statecheck.ExpectKnownValue("data.uptimekuma_maintenance.test", tfjsonpath.New("title"), knownvalue.StringExact(title)),
				},
			},
		},
	})
}

func testAccMaintenanceDataSourceConfigByID(title string) string {
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

data "uptimekuma_maintenance" "test" {
  id = uptimekuma_maintenance.test.id
}
`, title)
}

func TestAccMaintenanceDataSource_NotFoundByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
