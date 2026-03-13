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

func TestAccNotificationOneChatResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOneChat")
	nameUpdated := acctest.RandomWithPrefix("NotificationOneChatUpdated")
	accessToken := "test-access-token-123"
	accessTokenUpdated := "test-access-token-456"
	receiverID := "user123"
	receiverIDUpdated := "group456"
	botID := "bot-001"
	botIDUpdated := "bot-002"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOneChatResourceConfig(
					name,
					accessToken,
					receiverID,
					botID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("receiver_id"),
						knownvalue.StringExact(receiverID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("bot_id"),
						knownvalue.StringExact(botID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationOneChatResourceConfig(
					nameUpdated,
					accessTokenUpdated,
					receiverIDUpdated,
					botIDUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("access_token"),
						knownvalue.StringExact(accessTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("receiver_id"),
						knownvalue.StringExact(receiverIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("bot_id"),
						knownvalue.StringExact(botIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_onechat.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_onechat.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"access_token"},
			},
		},
	})
}

func testAccNotificationOneChatResourceConfig(
	name string,
	accessToken string,
	receiverID string,
	botID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onechat" "test" {
  name         = %[1]q
  is_active    = true
  access_token = %[2]q
  receiver_id  = %[3]q
  bot_id       = %[4]q
}
`, name, accessToken, receiverID, botID)
}
