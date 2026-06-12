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

func TestAccNotificationMaxResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationMax")
	nameUpdated := acctest.RandomWithPrefix("NotificationMaxUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationMaxResourceConfig(
					name,
					"bot-token-123",
					"-12345",
					true,
					"Monitor: {{ name }}",
					"markdown",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("api_url"),
						knownvalue.StringExact("https://platform-api.max.ru"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact("bot-token-123"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact("-12345"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("use_template"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("template"),
						knownvalue.StringExact("Monitor: {{ name }}"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("template_format"),
						knownvalue.StringExact("markdown"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccNotificationMaxResourceConfig(
					nameUpdated,
					"bot-token-456",
					"-67890",
					false,
					"",
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("bot_token"),
						knownvalue.StringExact("bot-token-456"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact("-67890"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_max.test",
						tfjsonpath.New("use_template"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_max.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationMaxImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bot_token"},
			},
		},
	})
}

func testAccNotificationMaxImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_max.test"]
	return rs.Primary.Attributes["id"], nil
}

//nolint:revive // flag-parameter is fine for test helper function
func testAccNotificationMaxResourceConfig(
	name string, botToken string, chatID string, useTemplate bool, template string, templateFormat string,
) string {
	templateFields := ""
	if useTemplate {
		templateFields = fmt.Sprintf("\n  template        = %q\n  template_format = %q", template, templateFormat)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_max" "test" {
  name         = %[1]q
  is_active    = true
  bot_token    = %[2]q
  chat_id      = %[3]q
  use_template = %[4]t%[5]s
}
`, name, botToken, chatID, useTemplate, templateFields)
}
