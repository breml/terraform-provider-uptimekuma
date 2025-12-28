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

func TestAccNotificationMattermostResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationMattermost")
	nameUpdated := acctest.RandomWithPrefix("NotificationMattermostUpdated")
	webhookURL := "https://mattermost.example.com/hooks/xxx"
	webhookURLUpdated := "https://mattermost.example.com/hooks/yyy"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationMattermostResourceConfig(
					name,
					webhookURL,
					"test-bot",
					"#general",
					":robot_face:",
					"https://example.com/icon.png",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("test-bot"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("#general"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("icon_emoji"),
						knownvalue.StringExact(":robot_face:"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("icon_url"),
						knownvalue.StringExact("https://example.com/icon.png"),
					),
				},
			},
			{
				Config: testAccNotificationMattermostResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					"updated-bot",
					"#alerts",
					":bell:",
					"https://example.com/updated-icon.png",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("updated-bot"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("#alerts"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("icon_emoji"),
						knownvalue.StringExact(":bell:"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_mattermost.test",
						tfjsonpath.New("icon_url"),
						knownvalue.StringExact("https://example.com/updated-icon.png"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_mattermost.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationMattermostImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url"},
			},
		},
	})
}

func testAccNotificationMattermostResourceConfig(
	name string, webhookURL string, username string, channel string,
	iconEmoji string, iconURL string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_mattermost" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
  username    = %[3]q
  channel     = %[4]q
  icon_emoji  = %[5]q
  icon_url    = %[6]q
}
`, name, webhookURL, username, channel, iconEmoji, iconURL)
}

func testAccNotificationMattermostImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_mattermost.test"]
	return rs.Primary.Attributes["id"], nil
}
