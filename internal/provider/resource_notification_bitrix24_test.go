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

// TestAccNotificationBitrix24Resource tests the Bitrix24 notification resource.
func TestAccNotificationBitrix24Resource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBitrix24")
	nameUpdated := acctest.RandomWithPrefix("NotificationBitrix24Updated")
	webhookURL := "https://your-bitrix24-domain.bitrix24.com/rest/1/webhook-key"
	webhookURLUpdated := "https://your-bitrix24-domain.bitrix24.com/rest/1/webhook-key-updated"
	userID := "123"
	userIDUpdated := "456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBitrix24ResourceConfig(
					name,
					webhookURL,
					userID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("notification_user_id"),
						knownvalue.StringExact(userID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationBitrix24ResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					userIDUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("notification_user_id"),
						knownvalue.StringExact(userIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

// testAccNotificationBitrix24ResourceConfig returns a Terraform configuration for testing.
func testAccNotificationBitrix24ResourceConfig(
	name string,
	webhookURL string,
	userID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bitrix24" "test" {
  name                    = %[1]q
  is_active               = true
  webhook_url             = %[2]q
  notification_user_id    = %[3]q
}
`, name, webhookURL, userID)
}
