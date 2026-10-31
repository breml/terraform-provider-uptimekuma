package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationSignalgridDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSignalgrid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSignalgridDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_signalgrid.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_signalgrid.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_signalgrid.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.CompareValuePairs(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("id"),
						"data.uptimekuma_notification_signalgrid.by_name",
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						"uptimekuma_notification_signalgrid.test",
						tfjsonpath.New("id"),
						"data.uptimekuma_notification_signalgrid.by_id",
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccNotificationSignalgridDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signalgrid" "test" {
  name       = %[1]q
  client_key = "client-key-1"
  channel    = "alerts"
}

data "uptimekuma_notification_signalgrid" "by_name" {
  name = uptimekuma_notification_signalgrid.test.name
}

data "uptimekuma_notification_signalgrid" "by_id" {
  id = uptimekuma_notification_signalgrid.test.id
}
`, name)
}
