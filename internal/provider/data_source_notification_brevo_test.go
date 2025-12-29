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

func TestAccNotificationBrevoDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationBrevo")
	apiKey := "test_api_key_1234567890"
	toEmail := "alerts@example.com"
	fromEmail := "monitoring@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBrevoDataSourceConfig(
					name,
					apiKey,
					toEmail,
					fromEmail,
					"Alert Subject",
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_brevo.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_brevo.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationBrevoDataSourceConfig(
	name string,
	apiKey string,
	toEmail string,
	fromEmail string,
	subject string,
	fromName string,
) string {
	fromNameConfig := ""
	if fromName != "" {
		fromNameConfig = fmt.Sprintf("  from_name = %q\n", fromName)
	}

	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_brevo" "test" {
  name       = %[1]q
  is_active  = true
  api_key    = %[2]q
  to_email   = %[3]q
  from_email = %[4]q
  subject    = %[5]q
%[6]s}

data "uptimekuma_notification_brevo" "test" {
  name = uptimekuma_notification_brevo.test.name
}
`, name, apiKey, toEmail, fromEmail, subject, fromNameConfig)
}
