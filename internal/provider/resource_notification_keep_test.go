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

func TestAccNotificationKeepResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationKeep")
	nameUpdated := acctest.RandomWithPrefix("NotificationKeepUpdated")
	webhookURL := "https://api.keephq.dev/alerts/alert"
	webhookURLUpdated := "https://api.keephq.dev/alerts/alert-updated"
	apiKey := "test-api-key-12345"
	apiKeyUpdated := "test-api-key-67890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationKeepResourceConfig(
					name,
					webhookURL,
					apiKey,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationKeepResourceConfig(
					nameUpdated,
					webhookURLUpdated,
					apiKeyUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact(webhookURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_keep.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_keep.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"webhook_url", "api_key"},
			},
		},
	})
}

func testAccNotificationKeepResourceConfig(
	name string, webhookURL string, apiKey string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_keep" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
  api_key     = %[3]q
}
`, name, webhookURL, apiKey)
}
