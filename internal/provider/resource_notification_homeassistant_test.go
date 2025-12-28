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

func TestAccNotificationHomeAssistantResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationHomeAssistant")
	nameUpdated := acctest.RandomWithPrefix("NotificationHomeAssistantUpdated")
	url := "https://homeassistant.example.com"
	urlUpdated := "https://homeassistant.updated.com"
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
	tokenUpdated := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9Updated"
	service := "notify.mobile_app"
	serviceUpdated := "notify.email"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHomeAssistantResourceConfig(
					name,
					url,
					token,
					service,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("home_assistant_url"),
						knownvalue.StringExact(url),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("long_lived_access_token"),
						knownvalue.StringExact(token),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("notification_service"),
						knownvalue.StringExact(service),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationHomeAssistantResourceConfig(
					nameUpdated,
					urlUpdated,
					tokenUpdated,
					serviceUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("home_assistant_url"),
						knownvalue.StringExact(urlUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("long_lived_access_token"),
						knownvalue.StringExact(tokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("notification_service"),
						knownvalue.StringExact(serviceUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_homeassistant.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_homeassistant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationHomeAssistantResourceConfig(
	name string,
	homeAssistantURL string,
	longLivedAccessToken string,
	notificationService string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_homeassistant" "test" {
  name                      = %[1]q
  is_active                 = true
  home_assistant_url        = %[2]q
  long_lived_access_token   = %[3]q
  notification_service      = %[4]q
}
`, name, homeAssistantURL, longLivedAccessToken, notificationService)
}
