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

func TestAccNotificationCallMeBotResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationCallMeBot")
	nameUpdated := acctest.RandomWithPrefix("NotificationCallMeBotUpdated")
	endpointURL := "https://api.callmebot.com/whatsapp.php?phone=1234567890&text="
	endpointURLUpdated := "https://api.callmebot.com/telegram.php?token=123456&chat_id=789&text="

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationCallMeBotResourceConfig(name, endpointURL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_callmebot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_callmebot.test",
						tfjsonpath.New("endpoint"),
						knownvalue.StringExact(endpointURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_callmebot.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationCallMeBotResourceConfig(nameUpdated, endpointURLUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_callmebot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_callmebot.test",
						tfjsonpath.New("endpoint"),
						knownvalue.StringExact(endpointURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_callmebot.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_callmebot.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccNotificationCallMeBotResourceConfig(name string, endpoint string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_callmebot" "test" {
  name     = %[1]q
  is_active = true
  endpoint = %[2]q
}
`, name, endpoint)
}
