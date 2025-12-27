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

func TestAccNotificationFeishuResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFeishu")
	nameUpdated := acctest.RandomWithPrefix("NotificationFeishuUpdated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFeishuResourceConfig(
					name,
					"https://open.feishu.cn/open-apis/bot/v2/hook/abcdefg123456",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_feishu.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_feishu.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_feishu.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact("https://open.feishu.cn/open-apis/bot/v2/hook/abcdefg123456"),
					),
				},
			},
			{
				Config: testAccNotificationFeishuResourceConfig(
					nameUpdated,
					"https://open.feishu.cn/open-apis/bot/v2/hook/updated789012",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_feishu.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_feishu.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_feishu.test",
						tfjsonpath.New("webhook_url"),
						knownvalue.StringExact("https://open.feishu.cn/open-apis/bot/v2/hook/updated789012"),
					),
				},
			},
		},
	})
}

func testAccNotificationFeishuResourceConfig(name string, webhookURL string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_feishu" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = %[2]q
}
`, name, webhookURL)
}
