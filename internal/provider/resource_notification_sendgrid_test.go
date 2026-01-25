package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationSendgridResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSendgrid")
	nameUpdated := acctest.RandomWithPrefix("NotificationSendgridUpdated")
	apiKey := "SG.test-api-key-" + acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	apiKeyUpdated := "SG.test-api-key-updated-" + acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSendgridResourceConfig(
					name,
					apiKey,
					"alerts@example.com",
					"monitoring@example.com",
					"Uptime Alert",
					"",
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("to_email"),
						knownvalue.StringExact("alerts@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("from_email"),
						knownvalue.StringExact("monitoring@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("subject"),
						knownvalue.StringExact("Uptime Alert"),
					),
				},
			},
			{
				Config: testAccNotificationSendgridResourceConfig(
					nameUpdated,
					apiKeyUpdated,
					"newalerts@example.com",
					"newmonitoring@example.com",
					"Service Alert",
					"cc@example.com",
					"bcc@example.com",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("to_email"),
						knownvalue.StringExact("newalerts@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("from_email"),
						knownvalue.StringExact("newmonitoring@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("subject"),
						knownvalue.StringExact("Service Alert"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("cc_email"),
						knownvalue.StringExact("cc@example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("bcc_email"),
						knownvalue.StringExact("bcc@example.com"),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_sendgrid.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationSendgridImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}

func TestAccNotificationSendgridDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSendgrid")
	apiKey := "SG.test-api-key-" + acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSendgridDataSourceConfig(
					name,
					apiKey,
					"alerts@example.com",
					"monitoring@example.com",
					"Uptime Alert",
					"",
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_sendgrid.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSendgridResourceConfig(
	name string, apiKey string, toEmail string, fromEmail string, subject string,
	ccEmail string, bccEmail string,
) string {
	ccEmailLine := ""
	if ccEmail != "" {
		ccEmailLine = fmt.Sprintf("  cc_email   = %q\n", ccEmail)
	}

	bccEmailLine := ""
	if bccEmail != "" {
		bccEmailLine = fmt.Sprintf("  bcc_email  = %q\n", bccEmail)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_sendgrid" "test" {
  name       = %[1]q
  is_active  = true
  api_key    = %[2]q
  to_email   = %[3]q
  from_email = %[4]q
  subject    = %[5]q
%[6]s%[7]s}
`, name, apiKey, toEmail, fromEmail, subject, ccEmailLine, bccEmailLine)
}

func testAccNotificationSendgridDataSourceConfig(
	name string, apiKey string, toEmail string, fromEmail string, subject string,
	ccEmail string, bccEmail string,
) string {
	ccEmailLine := ""
	if ccEmail != "" {
		ccEmailLine = fmt.Sprintf("  cc_email   = %q\n", ccEmail)
	}

	bccEmailLine := ""
	if bccEmail != "" {
		bccEmailLine = fmt.Sprintf("  bcc_email  = %q\n", bccEmail)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_sendgrid" "test" {
  name       = %[1]q
  is_active  = true
  api_key    = %[2]q
  to_email   = %[3]q
  from_email = %[4]q
  subject    = %[5]q
%[6]s%[7]s}

data "uptimekuma_notification_sendgrid" "test" {
  name = uptimekuma_notification_sendgrid.test.name
}
`, name, apiKey, toEmail, fromEmail, subject, ccEmailLine, bccEmailLine)
}

func testAccNotificationSendgridImportStateID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["uptimekuma_notification_sendgrid.test"]
	if !ok {
		return "", errors.New("Not found: uptimekuma_notification_sendgrid.test")
	}

	return rs.Primary.Attributes["id"], nil
}
