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

func TestAccNotificationGTXMessagingResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationGTXMessaging")
	nameUpdated := acctest.RandomWithPrefix("NotificationGTXMessagingUpdated")
	apiKey := "test-api-key-123"
	apiKeyUpdated := "test-api-key-456"
	from := "SenderID"
	fromUpdated := "NewSenderID"
	to := "+1234567890"
	toUpdated := "+0987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGTXMessagingResourceConfig(name, apiKey, from, to),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("from"),
						knownvalue.StringExact(from),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("to"),
						knownvalue.StringExact(to),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationGTXMessagingResourceConfig(nameUpdated, apiKeyUpdated, fromUpdated, toUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("from"),
						knownvalue.StringExact(fromUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("to"),
						knownvalue.StringExact(toUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_gtxmessaging.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccNotificationGTXMessagingResourceConfig(
	name string, apiKey string, from string, to string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gtxmessaging" "test" {
  name      = %[1]q
  is_active = true
  api_key   = %[2]q
  from      = %[3]q
  to        = %[4]q
}
`, name, apiKey, from, to)
}
