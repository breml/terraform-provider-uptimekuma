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

func TestAccNotificationFlashDutyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFlashDuty")
	nameUpdated := acctest.RandomWithPrefix("NotificationFlashDutyUpdated")
	integrationKey := "https://api.flashduty.com/webhook/events/12345678901234567890"
	integrationKeyUpdated := "https://api.flashduty.com/webhook/events/09876543210987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFlashDutyResourceConfig(
					name,
					integrationKey,
					"Warning",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("integration_key"),
						knownvalue.StringExact(integrationKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("severity"),
						knownvalue.StringExact("Warning"),
					),
				},
			},
			{
				Config: testAccNotificationFlashDutyResourceConfig(
					nameUpdated,
					integrationKeyUpdated,
					"Critical",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("integration_key"),
						knownvalue.StringExact(integrationKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_flashduty.test",
						tfjsonpath.New("severity"),
						knownvalue.StringExact("Critical"),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_flashduty.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationFlashDutyResourceConfig(
	name string,
	integrationKey string,
	severity string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flashduty" "test" {
  name              = %[1]q
  is_active         = true
  integration_key   = %[2]q
  severity          = %[3]q
}
`, name, integrationKey, severity)
}
