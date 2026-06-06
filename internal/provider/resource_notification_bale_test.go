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

func TestAccNotificationBaleResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBale")
	nameUpdated := acctest.RandomWithPrefix("NotificationBaleUpdated")
	botToken := "123456:ABCDEF"
	botTokenUpdated := "654321:FEDCBA"
	chatID := "111222333"
	chatIDUpdated := "444555666"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBaleResourceConfig(name, botToken, chatID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact(botToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact(chatID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccNotificationBaleResourceConfig(nameUpdated, botTokenUpdated, chatIDUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact(botTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact(chatIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bale.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_bale.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bot_token"},
			},
		},
	})
}

func testAccNotificationBaleResourceConfig(name string, botToken string, chatID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bale" "test" {
  name      = %[1]q
  is_active = true
  bot_token = %[2]q
  chat_id   = %[3]q
}
`, name, botToken, chatID)
}
