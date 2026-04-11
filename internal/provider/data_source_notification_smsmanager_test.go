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

func TestAccNotificationSMSManagerDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSMSManager")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSManagerDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSMSManagerDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsmanager.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSMSManagerDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsmanager" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-123"
  numbers   = "+1234567890"
}

data "uptimekuma_notification_smsmanager" "test" {
  name = uptimekuma_notification_smsmanager.test.name
}
`, name)
}

func testAccNotificationSMSManagerDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsmanager" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-123"
  numbers   = "+1234567890"
}

data "uptimekuma_notification_smsmanager" "test" {
  id = uptimekuma_notification_smsmanager.test.id
}
`, name)
}
