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

func TestAccNotificationOpsgenieDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOpsgenie")
	apiKey := "test-api-key-123"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOpsgenieDataSourceConfig(name, apiKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_opsgenie.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_opsgenie.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_opsgenie.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationOpsgenieDataSourceConfig(name string, apiKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_opsgenie" "test" {
  name     = %[1]q
  is_active = true
  api_key  = %[2]q
  region   = "us"
  priority = 1
}

data "uptimekuma_notification_opsgenie" "by_name" {
  name = uptimekuma_notification_opsgenie.test.name
}

data "uptimekuma_notification_opsgenie" "by_id" {
  id = uptimekuma_notification_opsgenie.test.id
}
`, name, apiKey)
}
