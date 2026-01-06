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

func TestAccNotificationKookDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationKook")
	botToken := "1/MzAxMjk5NzA1OTMzMDAwMA=="
	guildID := "382941547624206336"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationKookDataSourceConfig(name, botToken, guildID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_kook.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_kook.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationKookDataSourceConfig(name string, botToken string, guildID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_kook" "test" {
  name      = %[1]q
  is_active = true
  bot_token = %[2]q
  guild_id  = %[3]q
}

data "uptimekuma_notification_kook" "test" {
  name = uptimekuma_notification_kook.test.name
}
`, name, botToken, guildID)
}
