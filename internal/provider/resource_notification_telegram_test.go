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

func TestAccNotificationTelegramResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTelegram")
	nameUpdated := acctest.RandomWithPrefix("NotificationTelegramUpdated")
	botToken := "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghi"
	botTokenUpdated := "987654321:XYZabcdefghijklmnopqrstuvwxyzABCDEFG"
	chatID := "123456789"
	chatIDUpdated := "987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTelegramResourceConfig(
					name,
					botToken,
					chatID,
					false,
					false,
					"",
					"HTML",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact(botToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact(chatID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("send_silently"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("protect_content"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("use_template"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("template_parse_mode"),
						knownvalue.StringExact("HTML"),
					),
				},
			},
			{
				Config: testAccNotificationTelegramResourceConfig(
					nameUpdated,
					botTokenUpdated,
					chatIDUpdated,
					true,
					true,
					"https://custom-telegram-api.example.com",
					"Markdown",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact(botTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact(chatIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("send_silently"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("protect_content"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("server_url"),
						knownvalue.StringExact("https://custom-telegram-api.example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telegram.test",
						tfjsonpath.New("template_parse_mode"),
						knownvalue.StringExact("Markdown"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_telegram.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationTelegramImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bot_token"},
			},
		},
	})
}

func testAccNotificationTelegramImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_telegram.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationTelegramResourceConfig(
	name string,
	botToken string,
	chatID string,
	sendSilently bool,
	protectContent bool,
	serverURL string,
	parseMode string,
) string {
	serverURLConfig := ""
	if serverURL != "" {
		serverURLConfig = fmt.Sprintf("\n  server_url        = %q", serverURL)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_telegram" "test" {
  name                = %[1]q
  is_active           = true
  bot_token           = %[2]q
  chat_id             = %[3]q
  send_silently       = %[4]t
  protect_content     = %[5]t%[6]s
  template_parse_mode = %[7]q
}
`, name, botToken, chatID, sendSilently, protectContent, serverURLConfig, parseMode)
}
