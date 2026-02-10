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

func TestAccNotificationGTXMessagingDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationGTXMessaging")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGTXMessagingDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationGTXMessagingDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_gtxmessaging.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationGTXMessagingDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gtxmessaging" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-123"
  from      = "SenderID"
  to        = "+1234567890"
}

data "uptimekuma_notification_gtxmessaging" "test" {
  name = uptimekuma_notification_gtxmessaging.test.name
}
`, name)
}

func testAccNotificationGTXMessagingDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gtxmessaging" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-123"
  from      = "SenderID"
  to        = "+1234567890"
}

data "uptimekuma_notification_gtxmessaging" "test" {
  id = uptimekuma_notification_gtxmessaging.test.id
}
`, name)
}
