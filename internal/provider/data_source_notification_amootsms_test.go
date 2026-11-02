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

func TestAccNotificationAmootSMSDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationAmootSMS")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationAmootSMSDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_amootsms.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_amootsms.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_amootsms.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.CompareValuePairs(
						"uptimekuma_notification_amootsms.test",
						tfjsonpath.New("id"),
						"data.uptimekuma_notification_amootsms.by_name",
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						"uptimekuma_notification_amootsms.test",
						tfjsonpath.New("id"),
						"data.uptimekuma_notification_amootsms.by_id",
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccNotificationAmootSMSDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_amootsms" "test" {
  name        = %[1]q
  mobiles     = "09120000000"
  api_token   = "token-1"
  line_number = "public"
}

data "uptimekuma_notification_amootsms" "by_name" {
  name = uptimekuma_notification_amootsms.test.name
}

data "uptimekuma_notification_amootsms" "by_id" {
  id = uptimekuma_notification_amootsms.test.id
}
`, name)
}
