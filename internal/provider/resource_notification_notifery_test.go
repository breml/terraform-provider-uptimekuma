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

func TestAccNotificationNotiferyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationNotifery")
	nameUpdated := acctest.RandomWithPrefix("NotificationNotiferyUpdated")
	apiKey := "test-api-key-12345"
	apiKeyUpdated := "updated-api-key-67890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNotiferyResourceConfig(
					name,
					apiKey,
					"Test Alert",
					"test-group",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("test-group"),
					),
				},
			},
			{
				Config: testAccNotificationNotiferyResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					"Updated Alert",
					"updated-group",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Updated Alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_notifery.test",
						tfjsonpath.New("group"),
						knownvalue.StringExact("updated-group"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_notifery.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationNotiferyResourceConfig(
	name string, apiKey string, title string, group string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_notifery" "test" {
  name      = %[1]q
  is_active = true
  api_key   = %[2]q
  title     = %[3]q
  group     = %[4]q
}
`, name, apiKey, title, group)
}
