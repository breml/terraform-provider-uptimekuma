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

func TestAccNotificationTechulusPushResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationTechulusPush")
	nameUpdated := acctest.RandomWithPrefix("NotificationTechulusPushUpdated")
	apiKey := "test-api-key-12345"
	apiKeyUpdated := "test-api-key-67890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTechulusPushResourceConfig(
					name,
					apiKey,
					"Alert Title",
					"default",
					"alerts",
					true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Alert Title"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("sound"),
						knownvalue.StringExact("default"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("alerts"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("time_sensitive"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationTechulusPushResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					"Updated Title",
					"alarm",
					"critical",
					false,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Updated Title"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("sound"),
						knownvalue.StringExact("alarm"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("channel"),
						knownvalue.StringExact("critical"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("time_sensitive"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_techuluspush.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationTechulusPushResourceConfig(
	name string, apiKey string, title string, sound string, channel string,
	timeSensitive bool,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_techuluspush" "test" {
  name           = %[1]q
  is_active      = true
  api_key        = %[2]q
  title          = %[3]q
  sound          = %[4]q
  channel        = %[5]q
  time_sensitive = %[6]t
}
`, name, apiKey, title, sound, channel, timeSensitive)
}
