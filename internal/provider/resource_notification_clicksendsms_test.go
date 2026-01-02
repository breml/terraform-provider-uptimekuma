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

func TestAccNotificationClicksendSmsResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationClicksendSms")
	nameUpdated := acctest.RandomWithPrefix("NotificationClicksendSmsUpdated")
	login := "test@example.com"
	loginUpdated := "updated@example.com"
	password := "testApiKey123"
	passwordUpdated := "updatedApiKey456"
	toNumber := "+61412345678"
	toNumberUpdated := "+61487654321"
	senderName := "TestSender"
	senderNameUpdated := "UpdatedSender"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationClicksendSmsResourceConfig(
					name,
					login,
					password,
					toNumber,
					senderName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(login),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationClicksendSmsResourceConfig(
					nameUpdated,
					loginUpdated,
					passwordUpdated,
					toNumberUpdated,
					senderNameUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(loginUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(passwordUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_clicksendsms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_clicksendsms.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"login", "password"},
			},
		},
	})
}

func testAccNotificationClicksendSmsResourceConfig(
	name string, login string, password string, toNumber string, senderName string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_clicksendsms" "test" {
  name        = %[1]q
  is_active   = true
  login       = %[2]q
  password    = %[3]q
  to_number   = %[4]q
  sender_name = %[5]q
}
`, name, login, password, toNumber, senderName)
}
