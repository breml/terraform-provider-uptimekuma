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

func TestAccNotificationTwilioDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("DataSourceTwilio")
	accountSID := "account_sid_placeholder"
	authToken := "auth_token_placeholder"
	toNumber := "+12025550123"
	fromNumber := "+12025550789"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTwilioDataSourceConfig(name, accountSID, authToken, toNumber, fromNumber),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_twilio.by_id",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_twilio.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_twilio.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_twilio.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationTwilioDataSourceConfig(
	name string, accountSID string, authToken string, toNumber string, fromNumber string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_twilio" "test" {
  name        = %[1]q
  is_active   = true
  account_sid = %[2]q
  auth_token  = %[3]q
  to_number   = %[4]q
  from_number = %[5]q
}

data "uptimekuma_notification_twilio" "by_id" {
  id = uptimekuma_notification_twilio.test.id
}

data "uptimekuma_notification_twilio" "by_name" {
  name = uptimekuma_notification_twilio.test.name
}
`, name, accountSID, authToken, toNumber, fromNumber)
}
