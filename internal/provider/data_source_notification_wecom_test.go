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

func TestAccNotificationWeComDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("DataSourceWeCom")
	botKey := "bot_key_placeholder"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWeComDataSourceConfig(name, botKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_wecom.by_id",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_wecom.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_wecom.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_wecom.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationWeComDataSourceConfig(name string, botKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wecom" "test" {
  name    = %[1]q
  is_active = true
  bot_key = %[2]q
}

data "uptimekuma_notification_wecom" "by_id" {
  id = uptimekuma_notification_wecom.test.id
}

data "uptimekuma_notification_wecom" "by_name" {
  name = uptimekuma_notification_wecom.test.name
}
`, name, botKey)
}
