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

func TestAccNotificationNextcloudTalkResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNextcloudTalk")
	nameUpdated := acctest.RandomWithPrefix("NotificationNextcloudTalkUpdated")
	host := "https://nextcloud.example.com"
	hostUpdated := "https://cloud.example.org"
	conversationToken := "abc123token"
	conversationTokenUpdated := "xyz789token"
	botSecret := "secret123"
	botSecretUpdated := "secret456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNextcloudTalkResourceConfig(
					name,
					host,
					conversationToken,
					botSecret,
					false,
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(host),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("conversation_token"),
						knownvalue.StringExact(conversationToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("bot_secret"),
						knownvalue.StringExact(botSecret),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("send_silent_up"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("send_silent_down"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				Config: testAccNotificationNextcloudTalkResourceConfig(
					nameUpdated,
					hostUpdated,
					conversationTokenUpdated,
					botSecretUpdated,
					true,
					true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("host"),
						knownvalue.StringExact(hostUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("conversation_token"),
						knownvalue.StringExact(conversationTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("bot_secret"),
						knownvalue.StringExact(botSecretUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("send_silent_up"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("send_silent_down"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_nextcloudtalk.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationNextcloudTalkResourceConfig(
	name string,
	host string,
	conversationToken string,
	botSecret string,
	sendSilentUp bool,
	sendSilentDown bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_nextcloudtalk" "test" {
  name               = %[1]q
  is_active          = true
  host               = %[2]q
  conversation_token = %[3]q
  bot_secret         = %[4]q
  send_silent_up     = %[5]t
  send_silent_down   = %[6]t
}
`, name, host, conversationToken, botSecret, sendSilentUp, sendSilentDown)
}
