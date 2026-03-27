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

func TestAccNotificationSerwersmsResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSerwersms")
	nameUpdated := acctest.RandomWithPrefix("NotificationSerwersmsUpdated")
	username := "testuser"
	usernameUpdated := "testuserupdated"
	password := "testpass123"
	passwordUpdated := "testpass456"
	phoneNumber := "+48501234567"
	phoneNumberUpdated := "+48507654321"
	senderName := "TestSender"
	senderNameUpdated := "UpdatedSender"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSerwersmsResourceConfig(
					name,
					username,
					password,
					phoneNumber,
					senderName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSerwersmsResourceConfig(
					nameUpdated,
					usernameUpdated,
					passwordUpdated,
					phoneNumberUpdated,
					senderNameUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(usernameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(passwordUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_serwersms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_serwersms.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"username", "password"},
			},
		},
	})
}

func testAccNotificationSerwersmsResourceConfig(
	name string, username string, password string, phoneNumber string, senderName string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_serwersms" "test" {
  name         = %[1]q
  is_active    = true
  username     = %[2]q
  password     = %[3]q
  phone_number = %[4]q
  sender_name  = %[5]q
}
`, name, username, password, phoneNumber, senderName)
}
