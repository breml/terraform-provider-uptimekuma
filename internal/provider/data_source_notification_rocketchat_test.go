package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationRocketChatDataSource(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("NotificationRocketChat")
	dataSourceName := resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRocketChatDataSourceConfig(resourceName, dataSourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.uptimekuma_notification_rocketchat.test",
						"id",
					),
					resource.TestCheckResourceAttr(
						"data.uptimekuma_notification_rocketchat.test",
						"name",
						dataSourceName,
					),
				),
			},
		},
	})
}

func testAccNotificationRocketChatDataSourceConfig(resourceName string, _ string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_rocketchat" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://rocket.example.com/hooks/uid/token"
  username    = "rocket-bot"
  icon_emoji  = ":rocket:"
  channel     = "general"
  button      = "Visit Site"
}

data "uptimekuma_notification_rocketchat" "test" {
  name = uptimekuma_notification_rocketchat.test.name
}
`, resourceName)
}
