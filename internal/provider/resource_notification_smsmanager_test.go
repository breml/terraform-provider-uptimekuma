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

func TestAccNotificationSMSManagerResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSManager")
	nameUpdated := acctest.RandomWithPrefix("NotificationSMSManagerUpdated")
	apiKey := "test-api-key-123"
	apiKeyUpdated := "test-api-key-456"
	numbers := "+1234567890"
	numbersUpdated := "+0987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSManagerResourceConfig(
					name,
					apiKey,
					numbers,
					"sms",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("numbers"),
						knownvalue.StringExact(numbers),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("message_type"),
						knownvalue.StringExact("sms"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSMSManagerResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					numbersUpdated,
					"sms",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("numbers"),
						knownvalue.StringExact(numbersUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("message_type"),
						knownvalue.StringExact("sms"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_smsmanager.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationSMSManagerResourceConfig(
	name string, apiKey string, numbers string, messageType string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsmanager" "test" {
  name         = %[1]q
  is_active    = true
  api_key      = %[2]q
  numbers      = %[3]q
  message_type = %[4]q
}
`, name, apiKey, numbers, messageType)
}
