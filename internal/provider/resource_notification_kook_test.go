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

func TestAccNotificationKookResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationKook")
	nameUpdated := acctest.RandomWithPrefix("NotificationKookUpdated")
	botToken := "1/MzAxMjk5NzA1OTMxODAwMA=="
	botTokenUpdated := "1/MzAxMjk5NzA1OTMzMDAwMA=="
	guildID := "382941547624206336"
	guildIDUpdated := "382941547624206337"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationKookResourceConfig(
					name,
					botToken,
					guildID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact(botToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("guild_id"),
						knownvalue.StringExact(guildID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationKookResourceConfig(
					nameUpdated,
					botTokenUpdated,
					guildIDUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact(botTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("guild_id"),
						knownvalue.StringExact(guildIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_kook.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_kook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationKookResourceConfig(
	name string,
	botToken string,
	guildID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_kook" "test" {
  name      = %[1]q
  is_active = true
  bot_token = %[2]q
  guild_id  = %[3]q
}
`, name, botToken, guildID)
}
