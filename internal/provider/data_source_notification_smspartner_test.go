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

func TestAccNotificationSMSPartnerDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSMSPartner")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSPartnerDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smspartner.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSMSPartnerDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smspartner.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSMSPartnerDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smspartner" "test" {
  name         = %[1]q
  is_active    = true
  api_key      = "test_api_key"
  phone_number = "+33612345678"
}

data "uptimekuma_notification_smspartner" "test" {
  name = uptimekuma_notification_smspartner.test.name
}
`, name)
}

func testAccNotificationSMSPartnerDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smspartner" "test" {
  name         = %[1]q
  is_active    = true
  api_key      = "test_api_key"
  phone_number = "+33612345678"
}

data "uptimekuma_notification_smspartner" "test" {
  id = uptimekuma_notification_smspartner.test.id
}
`, name)
}
