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

func TestAccNotificationOctopushResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOctopush")
	nameUpdated := acctest.RandomWithPrefix("NotificationOctopushUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOctopushResourceConfig(
					name,
					"2",
					"test-api-key",
					"test-login",
					"+1234567890",
					"sms_premium",
					"TestSender",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("version"),
						knownvalue.StringExact("2"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("test-api-key"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("login"),
						knownvalue.StringExact("test-login"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("phone_number"),
						knownvalue.StringExact("+1234567890"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("sms_type"),
						knownvalue.StringExact("sms_premium"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("sender_name"),
						knownvalue.StringExact("TestSender"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationOctopushResourceConfig(
					nameUpdated,
					"1",
					"updated-api-key",
					"updated-login",
					"+9876543210",
					"sms_low_cost",
					"UpdatedSender",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("version"),
						knownvalue.StringExact("1"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("dm_api_key"),
						knownvalue.StringExact("updated-api-key"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("dm_login"),
						knownvalue.StringExact("updated-login"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("dm_phone_number"),
						knownvalue.StringExact("+9876543210"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("dm_sms_type"),
						knownvalue.StringExact("sms_low_cost"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_octopush.test",
						tfjsonpath.New("dm_sender_name"),
						knownvalue.StringExact("UpdatedSender"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_octopush.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationOctopushResourceConfig(
	name string, version string, apiKey string, login string, phoneNumber string,
	smsType string, senderName string,
) string {
	if version == "2" {
		return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_octopush" "test" {
  name          = %[1]q
  is_active     = true
  version       = %[2]q
  api_key       = %[3]q
  login         = %[4]q
  phone_number  = %[5]q
  sms_type      = %[6]q
  sender_name   = %[7]q
}
`, name, version, apiKey, login, phoneNumber, smsType, senderName)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_octopush" "test" {
  name            = %[1]q
  is_active       = true
  version         = %[2]q
  dm_api_key      = %[3]q
  dm_login        = %[4]q
  dm_phone_number = %[5]q
  dm_sms_type     = %[6]q
  dm_sender_name  = %[7]q
}
`, name, version, apiKey, login, phoneNumber, smsType, senderName)
}
