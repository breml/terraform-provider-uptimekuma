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

func TestAccNotificationWhapiResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWhapi")
	nameUpdated := acctest.RandomWithPrefix("NotificationWhapiUpdated")
	apiURL := "https://gate.whapi.cloud"
	apiURLUpdated := "https://gate-new.whapi.cloud"
	authToken := "test-auth-token-123"
	authTokenUpdated := "test-auth-token-456"
	recipient := "1234567890"
	recipientUpdated := "0987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWhapiResourceConfig(
					name,
					apiURL,
					authToken,
					recipient,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("api_url"),
						knownvalue.StringExact(apiURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipient),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationWhapiResourceConfig(
					nameUpdated,
					apiURLUpdated,
					authTokenUpdated,
					recipientUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("api_url"),
						knownvalue.StringExact(apiURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipientUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_whapi.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_whapi.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationWhapiResourceConfig(
	name string,
	apiURL string,
	authToken string,
	recipient string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_whapi" "test" {
  name       = %[1]q
  is_active  = true
  api_url    = %[2]q
  auth_token = %[3]q
  recipient  = %[4]q
}
`, name, apiURL, authToken, recipient)
}
