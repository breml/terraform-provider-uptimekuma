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

func TestAccNotificationSMSPlanetResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSPlanet")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMSPlanetUpdated")
	apiToken := "test-api-token-12345"
	apiTokenUpdated := "test-api-token-67890"
	phoneNumbers := "+48123456789"
	phoneNumbersUpdated := "+48987654321"
	senderName := "UptimeKuma"
	senderNameUpdated := "Monitoring"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSPlanetResourceConfig(
					name,
					apiToken,
					phoneNumbers,
					senderName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("api_token"),
						knownvalue.StringExact(apiToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("phone_numbers"),
						knownvalue.StringExact(phoneNumbers),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSMSPlanetResourceConfig(
					nameUpdated,
					apiTokenUpdated,
					phoneNumbersUpdated,
					senderNameUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("api_token"),
						knownvalue.StringExact(apiTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("phone_numbers"),
						knownvalue.StringExact(phoneNumbersUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact(senderNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_smsplanet.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_token"},
			},
		},
	})
}

func testAccNotificationSMSPlanetResourceConfig(
	name string,
	apiToken string,
	phoneNumbers string,
	senderName string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsplanet" "test" {
  name          = %[1]q
  is_active     = true
  api_token     = %[2]q
  phone_numbers = %[3]q
  sender_name   = %[4]q
}
`, name, apiToken, phoneNumbers, senderName)
}
