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

func TestAccNotificationSMSCDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSC")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSCDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSMSCDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSMSCDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsc" "test" {
  name        = %[1]q
  is_active   = true
  login       = "testuser"
  password    = "testpass123"
  to_number   = "77123456789"
  sender_name = "Uptime"
  translit    = "0"
}

data "uptimekuma_notification_smsc" "test" {
  name = uptimekuma_notification_smsc.test.name
}
`, name)
}

func testAccNotificationSMSCDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsc" "test" {
  name        = %[1]q
  is_active   = true
  login       = "testuser"
  password    = "testpass123"
  to_number   = "77123456789"
  sender_name = "Uptime"
  translit    = "0"
}

data "uptimekuma_notification_smsc" "test" {
  id = uptimekuma_notification_smsc.test.id
}
`, name)
}
