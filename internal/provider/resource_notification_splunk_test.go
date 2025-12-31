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

func TestAccNotificationSplunkResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSplunk")
	nameUpdated := acctest.RandomWithPrefix("NotificationSplunkUpdated")
	restURL := "https://api.victorops.com/api/v2"
	restURLUpdated := "https://api.victorops.com/api/v3"
	severity := "critical"
	severityUpdated := "warning"
	autoResolve := "resolve"
	autoResolveUpdated := "no_resolve"
	integrationKey := "integration_key_12345"
	integrationKeyUpdated := "integration_key_67890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSplunkResourceConfig(
					name,
					restURL,
					severity,
					autoResolve,
					integrationKey,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("rest_url"),
						knownvalue.StringExact(restURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("severity"),
						knownvalue.StringExact(severity),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("auto_resolve"),
						knownvalue.StringExact(autoResolve),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("integration_key"),
						knownvalue.StringExact(integrationKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationSplunkResourceConfig(
					nameUpdated,
					restURLUpdated,
					severityUpdated,
					autoResolveUpdated,
					integrationKeyUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("rest_url"),
						knownvalue.StringExact(restURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("severity"),
						knownvalue.StringExact(severityUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("auto_resolve"),
						knownvalue.StringExact(autoResolveUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("integration_key"),
						knownvalue.StringExact(integrationKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_splunk.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_splunk.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationSplunkImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rest_url", "integration_key"},
			},
		},
	})
}

// testAccNotificationSplunkImportStateID extracts the resource ID for import testing.
func testAccNotificationSplunkImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_splunk.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationSplunkResourceConfig(
	name string,
	restURL string,
	severity string,
	autoResolve string,
	integrationKey string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_splunk" "test" {
  name              = %[1]q
  is_active         = true
  rest_url          = %[2]q
  severity          = %[3]q
  auto_resolve      = %[4]q
  integration_key   = %[5]q
}
`, name, restURL, severity, autoResolve, integrationKey)
}
