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

func TestAccNotificationSMSPartnerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSPartner")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMSPartnerUpdated")
	apiKey := "test_api_key_1"
	apiKeyUpdated := "test_api_key_2"
	phoneNumber := "+33612345678"
	phoneNumberUpdated := "+33698765432"
	senderName := "TestSender"
	senderNameUpdated := "UpdatedSender"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSPartnerResourceConfig(
					name,
					apiKey,
					phoneNumber,
					senderName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSMSPartnerResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					phoneNumberUpdated,
					senderNameUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact(phoneNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smspartner.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_smspartner.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationSMSPartnerResourceConfig(
	name string,
	apiKey string,
	phoneNumber string,
	senderName string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smspartner" "test" {
  name         = %[1]q
  is_active    = true
  api_key      = %[2]q
  phone_number = %[3]q
  sender_name  = %[4]q
}
`, name, apiKey, phoneNumber, senderName)
}
