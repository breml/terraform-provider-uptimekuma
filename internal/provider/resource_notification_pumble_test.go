package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationPumbleResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPumble")
	nameUpdated := acctest.RandomWithPrefix("NotificationPumbleUpdated")
	webhookURL := "https://api.pumble.com/workspaces/WORKSPACE_ID/incomingWebhooks/WEBHOOK_ID"
	webhookURLUpdated := "https://api.pumble.com/workspaces/WORKSPACE_ID/incomingWebhooks/WEBHOOK_ID_UPDATED"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPumbleResourceConfig(name, webhookURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pumble.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pumble.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pumble.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPumbleResourceConfig(nameUpdated, webhookURLUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pumble.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pumble.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pumble.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_pumble.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationPumbleImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url"},
			},
		},
	})
}

func testAccNotificationPumbleResourceConfig(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pumble" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
}
`, name, webhookURL)
}

func testAccNotificationPumbleImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_pumble.test"]
	return rs.Primary.Attributes["id"], nil
}
