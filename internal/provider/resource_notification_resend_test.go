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

func TestAccNotificationResendResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationResend")
	nameUpdated := acctest.RandomWithPrefix("NotificationResendUpdated")
	apiKey := "re_test_" + acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	apiKeyUpdated := "re_test_updated_" + acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Minimal config: verify optional fields default to null.
			{
				Config: testAccNotificationResendResourceConfigMinimal(
					name,
					apiKey,
					"monitoring@example.com",
					"alerts@example.com",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("is_default"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("apply_existing"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("from_email"),
						knownvalue.StringExact("monitoring@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("to_email"),
						knownvalue.StringExact("alerts@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("from_name"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("subject"),
						knownvalue.Null(),
					),
				},
			},
			// Full config: set optional fields.
			{
				Config: testAccNotificationResendResourceConfig(
					name,
					apiKey,
					"monitoring@example.com",
					"Uptime Kuma",
					"alerts@example.com",
					"Uptime Alert",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("from_email"),
						knownvalue.StringExact("monitoring@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("from_name"),
						knownvalue.StringExact("Uptime Kuma"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("to_email"),
						knownvalue.StringExact("alerts@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("subject"),
						knownvalue.StringExact("Uptime Alert"),
					),
				},
			},
			// Update config: change values.
			{
				Config: testAccNotificationResendResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					"newmonitoring@example.com",
					"Uptime Kuma Updated",
					"newalerts@example.com,second@example.com",
					"Service Alert",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("from_email"),
						knownvalue.StringExact("newmonitoring@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("from_name"),
						knownvalue.StringExact("Uptime Kuma Updated"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("to_email"),
						knownvalue.StringExact("newalerts@example.com,second@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_resend.test",
						tfjsonpath.New("subject"),
						knownvalue.StringExact("Service Alert"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_resend.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func testAccNotificationResendResourceConfigMinimal(
	name string, apiKey string, fromEmail string, toEmail string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_resend" "test" {
  name       = %[1]q
  is_active  = true
  api_key    = %[2]q
  from_email = %[3]q
  to_email   = %[4]q
}
`, name, apiKey, fromEmail, toEmail)
}

func testAccNotificationResendResourceConfig(
	name string, apiKey string, fromEmail string, fromName string, toEmail string, subject string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_resend" "test" {
  name       = %[1]q
  is_active  = true
  api_key    = %[2]q
  from_email = %[3]q
  from_name  = %[4]q
  to_email   = %[5]q
  subject    = %[6]q
}
`, name, apiKey, fromEmail, fromName, toEmail, subject)
}
