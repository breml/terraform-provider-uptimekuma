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

func TestAccNotificationSendgridDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSendgrid")
	apiKey := "SG.test-api-key-" + acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
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
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_sendgrid.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_sendgrid.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationSendgridDataSourceConfig(
	name string, apiKey string, toEmail string, fromEmail string, subject string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_sendgrid" "test" {
  name       = %[1]q
  is_active  = true
  api_key    = %[2]q
  to_email   = %[3]q
  from_email = %[4]q
  subject    = %[5]q
}

data "uptimekuma_notification_sendgrid" "by_id" {
  id = uptimekuma_notification_sendgrid.test.id
}

data "uptimekuma_notification_sendgrid" "by_name" {
  name = uptimekuma_notification_sendgrid.test.name
}
`, name, apiKey, toEmail, fromEmail, subject)
}
