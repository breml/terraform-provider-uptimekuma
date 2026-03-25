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

func TestAccNotificationPumbleDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPumble")
	webhookURL := "https://api.pumble.com/workspaces/WORKSPACE_ID/incomingWebhooks/WEBHOOK_ID"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPumbleDataSourceConfig(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pumble.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pumble.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pumble.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPumbleDataSourceConfig(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pumble" "test" {
  name        = %[1]q
  webhook_url = %[2]q
}

data "uptimekuma_notification_pumble" "by_name" {
  name = uptimekuma_notification_pumble.test.name
}

data "uptimekuma_notification_pumble" "by_id" {
  id = uptimekuma_notification_pumble.test.id
}
`, name, webhookURL)
}
