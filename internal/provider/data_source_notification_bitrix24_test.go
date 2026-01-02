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

// TestAccNotificationBitrix24DataSource tests the Bitrix24 notification data source.
func TestAccNotificationBitrix24DataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestBitrix24Notification")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBitrix24DataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationBitrix24DataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bitrix24.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

// testAccNotificationBitrix24DataSourceConfig returns a Terraform configuration for testing by name.
func testAccNotificationBitrix24DataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bitrix24" "test" {
  name                    = %[1]q
  is_active               = true
  webhook_url             = "https://your-bitrix24-domain.bitrix24.com/rest/1/webhook-key"
  notification_user_id    = "123"
}

data "uptimekuma_notification_bitrix24" "test" {
  name = uptimekuma_notification_bitrix24.test.name
}
`, name)
}

// testAccNotificationBitrix24DataSourceConfigByID returns a Terraform configuration for testing by ID.
func testAccNotificationBitrix24DataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bitrix24" "test" {
  name                    = %[1]q
  is_active               = true
  webhook_url             = "https://your-bitrix24-domain.bitrix24.com/rest/1/webhook-key"
  notification_user_id    = "123"
}

data "uptimekuma_notification_bitrix24" "test" {
  id = uptimekuma_notification_bitrix24.test.id
}
`, name)
}
