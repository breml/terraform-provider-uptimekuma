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

func TestAccNotificationSMSCResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSC")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMSCUpdated")
	login := "testuser"
	loginUpdated := "testuserupdated"
	password := "testpass123"
	passwordUpdated := "testpass456"
	toNumber := "77123456789"
	toNumberUpdated := "77987654321"
	senderName := "Uptime"
	senderNameUpdated := "UptimeUpdated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSCResourceConfig(
					name,
					login,
					password,
					toNumber,
					senderName,
					"1",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(login),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("translit"),
						knownvalue.StringExact("1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSMSCResourceConfig(
					nameUpdated,
					loginUpdated,
					passwordUpdated,
					toNumberUpdated,
					senderNameUpdated,
					"0",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(loginUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(passwordUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("translit"),
						knownvalue.StringExact("0"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsc.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_smsc.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccNotificationSMSCResourceConfig(
	name string, login string, password string, toNumber string, senderName string,
	translit string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsc" "test" {
  name        = %[1]q
  is_active   = true
  login       = %[2]q
  password    = %[3]q
  to_number   = %[4]q
  sender_name = %[5]q
  translit    = %[6]q
}
`, name, login, password, toNumber, senderName, translit)
}
