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

func TestAccNotificationPromoSMSResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPromoSMS")
	nameUpdated := acctest.RandomWithPrefix("NotificationPromoSMSUpdated")
	login := "testuser"
	loginUpdated := "testuserupdated"
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
				Config: testAccNotificationPromoSMSResourceConfig(
					name,
					login,
					password,
					phoneNumber,
					senderName,
					"1",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(login),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(password),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("sms_type"),
						knownvalue.StringExact("1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("allow_long_sms"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPromoSMSResourceConfig(
					nameUpdated,
					loginUpdated,
					passwordUpdated,
					phoneNumberUpdated,
					senderNameUpdated,
					"0",
					true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact(loginUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("password"),
						knownvalue.StringExact(passwordUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("sms_type"),
						knownvalue.StringExact("0"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("allow_long_sms"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_promosms.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_promosms.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccNotificationPromoSMSResourceConfig(
	name string, login string, password string, phoneNumber string, senderName string,
	smsType string, allowLongSMS bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_promosms" "test" {
  name           = %[1]q
  is_active      = true
  login          = %[2]q
  password       = %[3]q
  phone_number   = %[4]q
  sender_name    = %[5]q
  sms_type       = %[6]q
  allow_long_sms = %[7]t
}
`, name, login, password, phoneNumber, senderName, smsType, allowLongSMS)
}
