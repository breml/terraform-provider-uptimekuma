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

func TestAccNotificationThreemaResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationThreema")
	nameUpdated := acctest.RandomWithPrefix("NotificationThreemaUpdated")
	senderIdentity := "TESTID123"
	senderIdentityUpdated := "TESTID456"
	secret := "testsecret123456789"
	secretUpdated := "testsecret987654321"
	recipient := "john@example.com"
	recipientUpdated := "+41234567890"
	recipientType := "email"
	recipientTypeUpdated := "phone"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationThreemaResourceConfig(
					name,
					senderIdentity,
					secret,
					recipient,
					recipientType,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("sender_identity"),
						knownvalue.StringExact(senderIdentity),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("secret"),
						knownvalue.StringExact(secret),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipient),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("recipient_type"),
						knownvalue.StringExact(recipientType),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationThreemaResourceConfig(
					nameUpdated,
					senderIdentityUpdated,
					secretUpdated,
					recipientUpdated,
					recipientTypeUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("sender_identity"),
						knownvalue.StringExact(senderIdentityUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("secret"),
						knownvalue.StringExact(secretUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipientUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("recipient_type"),
						knownvalue.StringExact(recipientTypeUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_threema.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_threema.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNotificationThreemaDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationThreema")
	senderIdentity := "TESTID123"
	secret := "testsecret123456789"
	recipient := "john@example.com"
	recipientType := "email"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationThreemaDataSourceConfig(
					name,
					senderIdentity,
					secret,
					recipient,
					recipientType,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_threema.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationThreemaResourceConfig(
	name string,
	senderIdentity string,
	secret string,
	recipient string,
	recipientType string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_threema" "test" {
  name               = %[1]q
  is_active          = true
  sender_identity    = %[2]q
  secret             = %[3]q
  recipient          = %[4]q
  recipient_type     = %[5]q
}
`, name, senderIdentity, secret, recipient, recipientType)
}

func testAccNotificationThreemaDataSourceConfig(
	name string,
	senderIdentity string,
	secret string,
	recipient string,
	recipientType string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_threema" "test" {
  name               = %[1]q
  is_active          = true
  sender_identity    = %[2]q
  secret             = %[3]q
  recipient          = %[4]q
  recipient_type     = %[5]q
}

data "uptimekuma_notification_threema" "test" {
  name = uptimekuma_notification_threema.test.name
}
`, name, senderIdentity, secret, recipient, recipientType)
}
