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

func TestAccMaintenanceStatusPagesDataSource(t *testing.T) {
	maintenanceTitle := acctest.RandomWithPrefix("TestMaintenance")
	statusPageSlug1 := acctest.RandomWithPrefix("test-status-1")
	statusPageSlug2 := acctest.RandomWithPrefix("test-status-2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMaintenanceStatusPagesDataSourceConfig(
					maintenanceTitle,
					statusPageSlug1,
					statusPageSlug2,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_maintenance_status_pages.test",
						tfjsonpath.New("status_page_ids"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
		},
	})
}

func testAccMaintenanceStatusPagesDataSourceConfig(
	maintenanceTitle string,
	statusPageSlug1 string,
	statusPageSlug2 string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_maintenance" "test" {
  title    = %[1]q
  strategy = "manual"
}

resource "uptimekuma_status_page" "test1" {
  slug  = %[2]q
  title = "Test Status Page 1"
}

resource "uptimekuma_status_page" "test2" {
  slug  = %[3]q
  title = "Test Status Page 2"
}

resource "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id  = uptimekuma_maintenance.test.id
  status_page_ids = [
    uptimekuma_status_page.test1.id,
    uptimekuma_status_page.test2.id,
  ]
}

data "uptimekuma_maintenance_status_pages" "test" {
  maintenance_id = uptimekuma_maintenance.test.id
  depends_on     = [uptimekuma_maintenance_status_pages.test]
}
`, maintenanceTitle, statusPageSlug1, statusPageSlug2)
}
