package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationPushoverResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPushover")
	nameUpdated := acctest.RandomWithPrefix("NotificationPushoverUpdated")
	userKey := "test-user-key-123"
	userKeyUpdated := "test-user-key-456"
	appToken := "test-app-token-789"
	appTokenUpdated := "test-app-token-012"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushoverResourceConfig(name, userKey, appToken),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("user_key"),
						knownvalue.StringExact(userKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("app_token"),
						knownvalue.StringExact(appToken),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPushoverResourceConfig(nameUpdated, userKeyUpdated, appTokenUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("user_key"),
						knownvalue.StringExact(userKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("app_token"),
						knownvalue.StringExact(appTokenUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pushover.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_pushover.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationPushoverImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_key", "app_token"},
			},
		},
	})
}

func testAccNotificationPushoverResourceConfig(name string, userKey string, appToken string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushover" "test" {
  name      = %[1]q
  is_active = true
  user_key  = %[2]q
  app_token = %[3]q
}
`, name, userKey, appToken)
}

func testAccNotificationPushoverImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_pushover.test"]
	return rs.Primary.Attributes["id"], nil
}
