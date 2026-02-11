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

func TestAccNotificationHeiiOnCallResource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHeiiOnCall")
	nameUpdated := acctest.RandomWithPrefix("TestHeiiOnCallUpdated")
	apiKey := acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	apiKeyUpdated := acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	triggerID := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum)
	triggerIDUpdated := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNotificationHeiiOnCallResourceConfig(name, apiKey, triggerID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("trigger_id"),
						knownvalue.StringExact(triggerID),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccNotificationHeiiOnCallResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					triggerIDUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("trigger_id"),
						knownvalue.StringExact(triggerIDUpdated),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "uptimekuma_notification_heiioncall.test",
				ImportState:       true,
				ImportStateVerify: true,
				// API key is sensitive and won't be returned from read
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationHeiiOnCallResourceConfig(name string, apiKey string, triggerID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_heiioncall" "test" {
  name       = %[1]q
  api_key    = %[2]q
  trigger_id = %[3]q
}
`, name, apiKey, triggerID)
}

func TestAccNotificationHeiiOnCallDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHeiiOnCall")
	apiKey := acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	triggerID := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read via name
			{
				Config: testAccNotificationHeiiOnCallDataSourceConfig(name, apiKey, triggerID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationHeiiOnCallDataSourceConfig(
	name string,
	apiKey string,
	triggerID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_heiioncall" "test" {
  name       = %[1]q
  api_key    = %[2]q
  trigger_id = %[3]q
}

data "uptimekuma_notification_heiioncall" "test" {
  name = uptimekuma_notification_heiioncall.test.name
}
`, name, apiKey, triggerID)
}
