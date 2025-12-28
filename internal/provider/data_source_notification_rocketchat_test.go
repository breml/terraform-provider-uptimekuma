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

func TestAccNotificationRocketChatDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationRocketChat")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRocketChatDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationRocketChatDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_rocketchat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationRocketChatDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_rocketchat" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://rocket.example.com/hooks/uid/token"
}

data "uptimekuma_notification_rocketchat" "test" {
  name = uptimekuma_notification_rocketchat.test.name
}
`, name)
}

func testAccNotificationRocketChatDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_rocketchat" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://rocket.example.com/hooks/uid/token"
}

data "uptimekuma_notification_rocketchat" "test" {
  id = uptimekuma_notification_rocketchat.test.id
}
`, name)
}
