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
	webhookURL := "https://discord.com/api/webhooks/123456789/abcdefghijklmnop"
	webhookURLUpdated := "https://discord.com/api/webhooks/987654321/zyxwvutsrqponmlk"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationDiscordResourceConfig(
					name,
					webhookURL,
					"Uptime Kuma",
					"text",
					"",
					"",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("Uptime Kuma"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("channel_type"),
						knownvalue.StringExact("text"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config: testAccNotificationDiscordResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					"Alert Bot",
					"forum",
					"thread-123",
					"🚨 Critical Alert",
					true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact("Alert Bot"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("channel_type"),
						knownvalue.StringExact("forum"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("thread_id"),
						knownvalue.StringExact("thread-123"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("prefix_message"),
						knownvalue.StringExact("🚨 Critical Alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_discord.test",
						tfjsonpath.New("disable_url"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationDiscordResourceConfig(
	name string, webhookURL string, username string, channelType string, threadID string, prefixMessage string, disableURL bool,
) string {
	// Build optional fields
	threadIDField := ""
	if threadID != "" {
		threadIDField = fmt.Sprintf("\n  thread_id = %q", threadID)
	}

	channelTypeField := ""
	if channelType != "" {
		channelTypeField = fmt.Sprintf("\n  channel_type = %q", channelType)
	}

	prefixMessageField := ""
	if prefixMessage != "" {
		prefixMessageField = fmt.Sprintf("\n  prefix_message = %q", prefixMessage)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_discord" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
  username    = %[3]q%[4]s%[5]s%[6]s
  disable_url = %[7]v
}
`, name, webhookURL, username, channelTypeField, threadIDField, prefixMessageField, disableURL)
}
