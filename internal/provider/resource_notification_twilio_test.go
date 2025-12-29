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

func TestAccNotificationTwilioResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTwilio")
	nameUpdated := acctest.RandomWithPrefix("NotificationTwilioUpdated")
	accountSID := "account_sid_placeholder"
	accountSIDUpdated := "account_sid_updated_placeholder"
	authToken := "auth_token_placeholder"
	authTokenUpdated := "auth_token_updated_placeholder"
	toNumber := "+12025550123"
	toNumberUpdated := "+12025550456"
	fromNumber := "+12025550789"
	fromNumberUpdated := "+12025550999"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTwilioResourceConfig(
					name,
					accountSID,
					authToken,
					toNumber,
					fromNumber,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("account_sid"),
						knownvalue.StringExact(accountSID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("from_number"),
						knownvalue.StringExact(fromNumber),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationTwilioResourceConfig(
					nameUpdated,
					accountSIDUpdated,
					authTokenUpdated,
					toNumberUpdated,
					fromNumberUpdated,
					`api_key = "test_api_key"`,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("account_sid"),
						knownvalue.StringExact(accountSIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("to_number"),
						knownvalue.StringExact(toNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("from_number"),
						knownvalue.StringExact(fromNumberUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact("test_api_key"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_twilio.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func testAccNotificationTwilioResourceConfig(
	name string, accountSID string, authToken string, toNumber string,
	fromNumber string, apiKey string,
) string {
	apiKeyConfig := ""
	if apiKey != "" {
		apiKeyConfig = fmt.Sprintf("  %s\n", apiKey)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_twilio" "test" {
  name           = %[1]q
  is_active      = true
  account_sid    = %[2]q
  auth_token     = %[3]q
  to_number      = %[4]q
  from_number    = %[5]q
%[6]s}
`, name, accountSID, authToken, toNumber, fromNumber, apiKeyConfig)
}
