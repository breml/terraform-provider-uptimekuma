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

func TestAccNotificationWPushResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWPush")
	nameUpdated := acctest.RandomWithPrefix("NotificationWPushUpdated")
	apiKey := "test-wpush-api-key-123"
	apiKeyUpdated := "test-wpush-api-key-456"
	channel := "test-channel"
	channelUpdated := "updated-channel"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWPushResourceConfig(
					name,
					apiKey,
					channel,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact(channel),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationWPushResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					channelUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact(channelUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wpush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_wpush.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationWPushResourceConfig(
	name string,
	apiKey string,
	channel string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wpush" "test" {
  name      = %[1]q
  is_active = true
  api_key   = %[2]q
  channel   = %[3]q
}
`, name, apiKey, channel)
}
