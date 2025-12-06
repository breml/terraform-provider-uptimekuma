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

func TestAccMaintenancesDataSource(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenancesDataSourceConfig(title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_maintenances.test", tfjsonpath.New("maintenances"), knownvalue.ListSizeExact(2)),
				},
			},
		},
	})
}

func testAccMaintenancesDataSourceConfig(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test1" {
  title       = "%[1]s-1"
  description = "First test maintenance"
  strategy    = "single"
  active      = true
  start_date  = "2025-12-31T10:00:00Z"
  end_date    = "2025-12-31T12:00:00Z"
  timezone    = "UTC"
}

resource "uptimekuma_maintenance" "test2" {
  title       = "%[1]s-2"
  description = "Second test maintenance"
  strategy    = "manual"
  active      = false
  timezone    = "UTC"
}

data "uptimekuma_maintenances" "test" {
  depends_on = [
    uptimekuma_maintenance.test1,
    uptimekuma_maintenance.test2,
  ]
}
`, title)
}

func TestAccMaintenancesDataSource_Empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenancesDataSourceConfigEmpty(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_maintenances.test", tfjsonpath.New("maintenances"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testAccMaintenancesDataSourceConfigEmpty() string {
	return providerConfig() + `
data "uptimekuma_maintenances" "test" {
}
`
}

func TestAccMaintenancesDataSource_MultipleStrategies(t *testing.T) {
	title := acctest.RandomWithPrefix("TestMaintenance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenancesDataSourceConfigMultipleStrategies(title),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_maintenances.test", tfjsonpath.New("maintenances"), knownvalue.ListSizeExact(3)),
				},
			},
		},
	})
}

func testAccMaintenancesDataSourceConfigMultipleStrategies(title string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "single" {
  title       = "%[1]s-single"
  description = "Single maintenance"
  strategy    = "single"
  active      = true
  start_date  = "2025-12-31T10:00:00Z"
  end_date    = "2025-12-31T12:00:00Z"
  timezone    = "UTC"
}

resource "uptimekuma_maintenance" "cron" {
  title            = "%[1]s-cron"
  description      = "Cron maintenance"
  strategy         = "cron"
  active           = true
  cron             = "0 2 * * *"
  duration_minutes = 120
  timezone         = "UTC"
}

resource "uptimekuma_maintenance" "interval" {
  title        = "%[1]s-interval"
  description  = "Interval maintenance"
  strategy     = "recurring-interval"
  active       = true
  interval_day = 7
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

data "uptimekuma_maintenances" "test" {
  depends_on = [
    uptimekuma_maintenance.single,
    uptimekuma_maintenance.cron,
    uptimekuma_maintenance.interval,
  ]
}
`, title)
}
