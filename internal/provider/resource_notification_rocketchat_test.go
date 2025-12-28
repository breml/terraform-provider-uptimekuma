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

func TestAccNotificationRocketChatResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationRocketChat")
	nameUpdated := acctest.RandomWithPrefix("NotificationRocketChatUpdated")
	webhookURL := "https://rocket.example.com/hooks/uid/token"
	webhookURLUpdated := "https://rocket.example.com/hooks/uid/token-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRocketChatResourceConfig(
					name,
					webhookURL,
					"rocket-bot",
					":rocket:",
					"general",
					"Visit Site",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("rocket-bot"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("icon_emoji"),
						knownvalue.StringExact(":rocket:"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("general"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("button"),
						knownvalue.StringExact("Visit Site"),
					),
				},
			},
			{
				Config: testAccNotificationRocketChatResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					"updated-rocket-bot",
					":bell:",
					"alerts",
					"Check Status",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("updated-rocket-bot"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("icon_emoji"),
						knownvalue.StringExact(":bell:"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("alerts"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("button"),
						knownvalue.StringExact("Check Status"),
					),
				},
			},
		},
	})
}

func testAccNotificationRocketChatResourceConfig(
	name string, webhookURL string, username string, iconEmoji string, channel string,
	button string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_rocketchat" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
  username    = %[3]q
  icon_emoji  = %[4]q
  channel     = %[5]q
  button      = %[6]q
}
`, name, webhookURL, username, iconEmoji, channel, button)
}
