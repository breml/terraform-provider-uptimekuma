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

func TestAccNotificationOneChatDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationOneChat")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOneChatDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_onechat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationOneChatDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_onechat.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationOneChatDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onechat" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "test-access-token"
  receiver_id  = "user123"
  bot_id       = "bot-001"
}

data "uptimekuma_notification_onechat" "test" {
  name = uptimekuma_notification_onechat.test.name
}
`, name)
}

func testAccNotificationOneChatDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onechat" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "test-access-token"
  receiver_id  = "user123"
  bot_id       = "bot-001"
}

data "uptimekuma_notification_onechat" "test" {
  id = uptimekuma_notification_onechat.test.id
}
`, name)
}
