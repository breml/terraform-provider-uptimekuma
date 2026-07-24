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

func TestAccNotificationTelnyxResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTelnyx")
	nameUpdated := acctest.RandomWithPrefix("NotificationTelnyxUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTelnyxResourceConfig(
					name,
					"KEY0000000000000000000000000000000",
					"40017a13-3f93-4d2d-b29e-1a000000000a",
					"+15550001111",
					"+15550002222",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("KEY0000000000000000000000000000000"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("messaging_profile_id"),
						knownvalue.StringExact("40017a13-3f93-4d2d-b29e-1a000000000a"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact("+15550001111"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact("+15550002222"),
					),
				},
			},
			{
				Config: testAccNotificationTelnyxResourceConfig(
					nameUpdated,
					"KEY1111111111111111111111111111111",
					"40017a13-3f93-4d2d-b29e-1a000000000b",
					"+15550003333",
					"+15550004444",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("KEY1111111111111111111111111111111"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("messaging_profile_id"),
						knownvalue.StringExact("40017a13-3f93-4d2d-b29e-1a000000000b"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact("+15550003333"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_telnyx.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact("+15550004444"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_telnyx.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationTelnyxResourceConfig(
	name string, apiKey string, messagingProfileID string, phoneNumber string, toNumber string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_telnyx" "test" {
  name                 = %[1]q
  is_active            = true
  api_key              = %[2]q
  messaging_profile_id = %[3]q
  phone_number         = %[4]q
  to_number            = %[5]q
}
`, name, apiKey, messagingProfileID, phoneNumber, toNumber)
}
