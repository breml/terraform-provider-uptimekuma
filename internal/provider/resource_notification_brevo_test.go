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

func TestAccNotificationBrevoResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBrevo")
	nameUpdated := acctest.RandomWithPrefix("NotificationBrevoUpdated")
	apiKey := "test_api_key_1234567890"
	apiKeyUpdated := "updated_api_key_0987654321"
	toEmail := "alerts@example.com"
	toEmailUpdated := "notifications@example.com"
	fromEmail := "monitoring@example.com"
	fromEmailUpdated := "uptime@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBrevoResourceConfig(
					name,
					apiKey,
					toEmail,
					fromEmail,
					"Monitoring Alert",
					"monitoring-group",
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("to_email"),
						knownvalue.StringExact(toEmail),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("from_email"),
						knownvalue.StringExact(fromEmail),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("subject"),
						knownvalue.StringExact("Monitoring Alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("from_name"),
						knownvalue.StringExact("monitoring-group"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationBrevoResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					toEmailUpdated,
					fromEmailUpdated,
					"Updated Alert",
					"updated-group",
					"cc@example.com",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("to_email"),
						knownvalue.StringExact(toEmailUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("from_email"),
						knownvalue.StringExact(fromEmailUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("subject"),
						knownvalue.StringExact("Updated Alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("from_name"),
						knownvalue.StringExact("updated-group"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("cc_email"),
						knownvalue.StringExact("cc@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_brevo.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_brevo.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNotificationBrevoResourceConfig(
	name string,
	apiKey string,
	toEmail string,
	fromEmail string,
	subject string,
	fromName string,
	ccEmail string,
) string {
	ccEmailConfig := ""
	if ccEmail != "" {
		ccEmailConfig = fmt.Sprintf("  cc_email   = %q\n", ccEmail)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_brevo" "test" {
  name       = %[1]q
  is_active  = true
  api_key    = %[2]q
  to_email   = %[3]q
  from_email = %[4]q
  subject    = %[5]q
  from_name  = %[6]q
%[7]s}
`, name, apiKey, toEmail, fromEmail, subject, fromName, ccEmailConfig)
}
