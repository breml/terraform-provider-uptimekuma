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

func TestAccNotificationCallMeBotDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationCallMeBot")
	endpoint := "https://api.callmebot.com/whatsapp.php?phone=1234567890&text="

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationCallMeBotDataSourceConfig(name, endpoint),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_callmebot.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_callmebot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationCallMeBotDataSourceConfig(name string, endpoint string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_callmebot" "test" {
  name     = %[1]q
  is_active = true
  endpoint = %[2]q
}

data "uptimekuma_notification_callmebot" "test" {
  name = uptimekuma_notification_callmebot.test.name
}
`, name, endpoint)
}
