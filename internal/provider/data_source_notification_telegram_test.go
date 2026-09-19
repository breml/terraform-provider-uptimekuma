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

func TestAccNotificationTelegramDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationTelegram")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTelegramDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_telegram.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_telegram.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationTelegramDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_telegram" "test" {
  name      = %[1]q
  is_active = true
  bot_token = "123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghi"
  chat_id   = "123456789"
}

data "uptimekuma_notification_telegram" "by_name" {
  name = uptimekuma_notification_telegram.test.name
}

data "uptimekuma_notification_telegram" "by_id" {
  id = uptimekuma_notification_telegram.test.id
}
`, name)
}
