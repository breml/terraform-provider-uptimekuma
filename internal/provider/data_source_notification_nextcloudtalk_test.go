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

func TestAccNotificationNextcloudTalkDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationNextcloudTalk")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNextcloudTalkDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationNextcloudTalkDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_nextcloudtalk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationNextcloudTalkDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_nextcloudtalk" "test" {
  name               = %[1]q
  is_active          = true
  host               = "https://nextcloud.example.com"
  conversation_token = "test-token-123"
  bot_secret         = "test-secret-456"
}

data "uptimekuma_notification_nextcloudtalk" "test" {
  name = uptimekuma_notification_nextcloudtalk.test.name
}
`, name)
}

func testAccNotificationNextcloudTalkDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_nextcloudtalk" "test" {
  name               = %[1]q
  is_active          = true
  host               = "https://nextcloud.example.com"
  conversation_token = "test-token-123"
  bot_secret         = "test-secret-456"
}

data "uptimekuma_notification_nextcloudtalk" "test" {
  id = uptimekuma_notification_nextcloudtalk.test.id
}
`, name)
}
