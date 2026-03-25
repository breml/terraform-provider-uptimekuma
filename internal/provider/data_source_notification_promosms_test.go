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

func TestAccNotificationPromoSMSDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPromoSMS")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPromoSMSDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_promosms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationPromoSMSDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_promosms.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPromoSMSDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_promosms" "test" {
  name           = %[1]q
  is_active      = true
  login          = "testuser"
  password       = "testpass123"
  phone_number   = "+48501234567"
  sender_name    = "TestSender"
  sms_type       = "1"
  allow_long_sms = false
}

data "uptimekuma_notification_promosms" "test" {
  name = uptimekuma_notification_promosms.test.name
}
`, name)
}

func testAccNotificationPromoSMSDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_promosms" "test" {
  name           = %[1]q
  is_active      = true
  login          = "testuser"
  password       = "testpass123"
  phone_number   = "+48501234567"
  sender_name    = "TestSender"
  sms_type       = "1"
  allow_long_sms = false
}

data "uptimekuma_notification_promosms" "test" {
  id = uptimekuma_notification_promosms.test.id
}
`, name)
}
