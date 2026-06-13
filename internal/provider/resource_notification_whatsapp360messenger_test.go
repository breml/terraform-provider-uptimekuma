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

func TestAccNotificationWhatsapp360messengerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWhatsapp360messenger")
	nameUpdated := acctest.RandomWithPrefix("NotificationWhatsapp360messengerUpdated")
	authToken := "auth-token-xxxxxxxx"
	authTokenUpdated := "auth-token-yyyyyyyy"
	recipient := "+15551234567"
	recipientUpdated := "+15557654321,+15551112222"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Minimal config: verify optional fields default to null.
			{
				Config: testAccNotificationWhatsapp360messengerResourceConfigMinimal(name, authToken, recipient),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipient),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("group_ids"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("group_id"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("use_template"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("template"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// Update scalar fields and set pointer-based optional fields.
			{
				Config: testAccNotificationWhatsapp360messengerResourceConfig(
					nameUpdated,
					authTokenUpdated,
					recipientUpdated,
					"120363012345678901",
					true,
					"{{ name }} is {{ status }}",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipientUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("group_id"),
						knownvalue.StringExact("120363012345678901"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("use_template"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("template"),
						knownvalue.StringExact("{{ name }} is {{ status }}"),
					),
				},
			},
			// Set the group_ids list and verify it round-trips correctly.
			{
				Config: testAccNotificationWhatsapp360messengerResourceConfigWithGroupIDs(
					nameUpdated,
					authTokenUpdated,
					recipientUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("group_ids"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("120363012345678901"),
							knownvalue.StringExact("120363098765432109"),
						}),
					),
				},
			},
			// Clear pointer/list-based optional fields and verify they return to null.
			{
				Config: testAccNotificationWhatsapp360messengerResourceConfigMinimal(
					nameUpdated, authTokenUpdated, recipientUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("group_ids"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("group_id"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("use_template"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whatsapp360messenger.test",
						tfjsonpath.New("template"),
						knownvalue.Null(),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_whatsapp360messenger.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_token"},
			},
		},
	})
}

func testAccNotificationWhatsapp360messengerResourceConfigMinimal(
	name string, authToken string, recipient string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whatsapp360messenger" "test" {
  name       = %[1]q
  is_active  = true
  auth_token = %[2]q
  recipient  = %[3]q
}
`, name, authToken, recipient)
}

func testAccNotificationWhatsapp360messengerResourceConfig(
	name string, authToken string, recipient string, groupID string, useTemplate bool, template string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whatsapp360messenger" "test" {
  name         = %[1]q
  is_active    = true
  auth_token   = %[2]q
  recipient    = %[3]q
  group_id     = %[4]q
  use_template = %[5]t
  template     = %[6]q
}
`, name, authToken, recipient, groupID, useTemplate, template)
}

func testAccNotificationWhatsapp360messengerResourceConfigWithGroupIDs(
	name string, authToken string, recipient string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whatsapp360messenger" "test" {
  name       = %[1]q
  is_active  = true
  auth_token = %[2]q
  recipient  = %[3]q
  group_ids  = ["120363012345678901", "120363098765432109"]
}
`, name, authToken, recipient)
}
