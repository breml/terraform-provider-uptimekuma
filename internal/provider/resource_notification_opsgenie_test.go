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

func TestAccNotificationOpsgenieResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOpsgenie")
	nameUpdated := acctest.RandomWithPrefix("NotificationOpsgenieUpdated")
	apiKey := "test-api-key-123"
	apiKeyUpdated := "test-api-key-456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOpsgenieResourceConfig(
					name,
					apiKey,
					"us",
					1,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("region"),
						knownvalue.StringExact("us"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("priority"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationOpsgenieResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					"eu",
					3,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("region"),
						knownvalue.StringExact("eu"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("priority"),
						knownvalue.Int64Exact(3),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_opsgenie.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_opsgenie.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationOpsgenieImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationOpsgenieResourceConfig(
	name string, apiKey string, region string, priority int64,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_opsgenie" "test" {
  name     = %[1]q
  is_active = true
  api_key  = %[2]q
  region   = %[3]q
  priority = %[4]d
}
`, name, apiKey, region, priority)
}

func testAccNotificationOpsgenieImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_opsgenie.test"]
	return rs.Primary.Attributes["id"], nil
}
