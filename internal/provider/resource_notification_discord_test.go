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

func TestAccNotificationDiscordResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationDiscord")
	nameUpdated := acctest.RandomWithPrefix("NotificationDiscordUpdated")
	webhookURL := "https://discordapp.com/api/webhooks/1234567890/XXXXXXXXXXXXXXXXXXXX"
	webhookURLUpdated := "https://discordapp.com/api/webhooks/1234567890/YYYYYYYYYYYYYYYYYYYY"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationDiscordResourceConfig(name, webhookURL, "test-bot", "text"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("webhook_url"), knownvalue.StringExact(webhookURL)),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("username"), knownvalue.StringExact("test-bot")),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("channel_type"), knownvalue.StringExact("text")),
				},
			},
			{
				Config: testAccNotificationDiscordResourceConfig(nameUpdated, webhookURLUpdated, "updated-bot", "announcement"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("webhook_url"), knownvalue.StringExact(webhookURLUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("username"), knownvalue.StringExact("updated-bot")),
					statecheck.ExpectKnownValue("uptimekuma_notification_discord.test", tfjsonpath.New("channel_type"), knownvalue.StringExact("announcement")),
				},
			},
		},
	})
}

func testAccNotificationDiscordResourceConfig(name, webhookURL, username, channelType string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_discord" "test" {
  name         = %[1]q
  is_active    = true
  webhook_url  = %[2]q
  username     = %[3]q
  channel_type = %[4]q
}
`, name, webhookURL, username, channelType)
}
