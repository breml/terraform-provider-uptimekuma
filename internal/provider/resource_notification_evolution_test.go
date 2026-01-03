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

func TestAccNotificationEvolutionResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationEvolution")
	nameUpdated := acctest.RandomWithPrefix("NotificationEvolutionUpdated")
	apiURL := "https://api.evolution.example.com"
	apiURLUpdated := "https://api2.evolution.example.com"
	instanceName := "testinstance"
	instanceNameUpdated := "updatedinstance"
	authToken := "testAuthToken123"
	authTokenUpdated := "updatedAuthToken456"
	recipient := "+551198765432"
	recipientUpdated := "+551234567890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationEvolutionResourceConfig(
					name,
					apiURL,
					instanceName,
					authToken,
					recipient,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("api_url"),
						knownvalue.StringExact(apiURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("instance_name"),
						knownvalue.StringExact(instanceName),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipient),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationEvolutionResourceConfig(
					nameUpdated,
					apiURLUpdated,
					instanceNameUpdated,
					authTokenUpdated,
					recipientUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("api_url"),
						knownvalue.StringExact(apiURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("instance_name"),
						knownvalue.StringExact(instanceNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("auth_token"),
						knownvalue.StringExact(authTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("recipient"),
						knownvalue.StringExact(recipientUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_evolution.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_evolution.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccNotificationEvolutionResourceConfig(
	name string, apiURL string, instanceName string, authToken string, recipient string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_evolution" "test" {
  name            = %[1]q
  is_active       = true
  api_url         = %[2]q
  instance_name   = %[3]q
  auth_token      = %[4]q
  recipient       = %[5]q
}
`, name, apiURL, instanceName, authToken, recipient)
}
