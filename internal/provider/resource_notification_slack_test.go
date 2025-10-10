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

func TestAccNotificationSlackResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSlack")
	nameUpdated := acctest.RandomWithPrefix("NotificationSlackUpdated")
	webhookURL := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSlackResourceConfig(name, webhookURL, "test-bot", ":robot_face:", "#general", true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("username"), knownvalue.StringExact("test-bot")),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("icon_emoji"), knownvalue.StringExact(":robot_face:")),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("channel"), knownvalue.StringExact("#general")),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("channel_notify"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccNotificationSlackResourceConfig(nameUpdated, webhookURL, "updated-bot", ":bell:", "#alerts", false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("name"), knownvalue.StringExact(nameUpdated)),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("is_active"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("username"), knownvalue.StringExact("updated-bot")),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("icon_emoji"), knownvalue.StringExact(":bell:")),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("channel"), knownvalue.StringExact("#alerts")),
					statecheck.ExpectKnownValue("uptimekuma_notification_slack.test", tfjsonpath.New("channel_notify"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccNotificationSlackResourceConfig(name, webhookURL, username, iconEmoji, channel string, channelNotify bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_slack" "test" {
  name           = %[1]q
  is_active      = true
  webhook_url    = %[2]q
  username       = %[3]q
  icon_emoji     = %[4]q
  channel        = %[5]q
  channel_notify = %[6]t
}
`, name, webhookURL, username, iconEmoji, channel, channelNotify)
}
