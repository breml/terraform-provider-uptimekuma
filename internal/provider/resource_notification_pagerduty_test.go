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

func TestAccNotificationPagerDutyResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPagerDuty")
	nameUpdated := acctest.RandomWithPrefix("NotificationPagerDutyUpdated")
	integrationURL := "https://events.pagerduty.com/integration/test/enqueue"
	integrationURLUpdated := "https://events.pagerduty.com/integration/test2/enqueue"
	integrationKey := "test-key-123456789"
	integrationKeyUpdated := "updated-key-987654321"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPagerDutyResourceConfig(
					name,
					integrationURL,
					integrationKey,
					"high",
					"resolved",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("integration_url"),
						knownvalue.StringExact(integrationURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("integration_key"),
						knownvalue.StringExact(integrationKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("priority"),
						knownvalue.StringExact("high"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("auto_resolve"),
						knownvalue.StringExact("resolved"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationPagerDutyResourceConfig(
					nameUpdated,
					integrationURLUpdated,
					integrationKeyUpdated,
					"critical",
					"triggered",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("integration_url"),
						knownvalue.StringExact(integrationURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("integration_key"),
						knownvalue.StringExact(integrationKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("priority"),
						knownvalue.StringExact("critical"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("auto_resolve"),
						knownvalue.StringExact("triggered"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_pagerduty.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationPagerDutyImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"integration_url", "integration_key"},
			},
		},
	})
}

// testAccNotificationPagerDutyImportStateID extracts the resource ID for import testing.
func testAccNotificationPagerDutyImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_pagerduty.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationPagerDutyResourceConfig(
	name string, integrationURL string, integrationKey string,
	priority string, autoResolve string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pagerduty" "test" {
  name              = %[1]q
  is_active         = true
  integration_url   = %[2]q
  integration_key   = %[3]q
  priority          = %[4]q
  auto_resolve      = %[5]q
}
`, name, integrationURL, integrationKey, priority, autoResolve)
}
