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

func TestAccNotificationNotifyAppDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationNotifyApp")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNotifyAppDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_notifyapp.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_notifyapp.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_notifyapp.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.CompareValuePairs(
						"uptimekuma_notification_notifyapp.test",
						tfjsonpath.New("id"),
						"data.uptimekuma_notification_notifyapp.by_name",
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						"uptimekuma_notification_notifyapp.test",
						tfjsonpath.New("id"),
						"data.uptimekuma_notification_notifyapp.by_id",
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccNotificationNotifyAppDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_notifyapp" "test" {
  name      = %[1]q
  device_id = "ABC12345"
  token     = "token-1"
}

data "uptimekuma_notification_notifyapp" "by_name" {
  name = uptimekuma_notification_notifyapp.test.name
}

data "uptimekuma_notification_notifyapp" "by_id" {
  id = uptimekuma_notification_notifyapp.test.id
}
`, name)
}
