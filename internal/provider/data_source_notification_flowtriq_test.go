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

func TestAccNotificationFlowtriqDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationFlowtriq")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFlowtriqDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_flowtriq.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_flowtriq.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_flowtriq.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationFlowtriqDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flowtriq" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://app.flowtriq.com/api/webhooks/0123456789abcdef"
  api_key     = "apikey0000000000"
}

data "uptimekuma_notification_flowtriq" "by_name" {
  name = uptimekuma_notification_flowtriq.test.name
}

data "uptimekuma_notification_flowtriq" "by_id" {
  id = uptimekuma_notification_flowtriq.test.id
}
`, name)
}
