package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationWeComResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWeCom")
	nameUpdated := acctest.RandomWithPrefix("NotificationWeComUpdated")
	botKey := "bot_key_placeholder"
	botKeyUpdated := "bot_key_updated_placeholder"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWeComResourceConfig(name, botKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wecom.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wecom.test",
						tfjsonpath.New("bot_key"),
						knownvalue.StringExact(botKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wecom.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationWeComResourceConfig(nameUpdated, botKeyUpdated),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wecom.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wecom.test",
						tfjsonpath.New("bot_key"),
						knownvalue.StringExact(botKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_wecom.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:            "uptimekuma_notification_wecom.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccNotificationWeComImportStateID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bot_key"},
			},
		},
	})
}

func testAccNotificationWeComImportStateID(s *terraform.State) (string, error) {
	rs := s.RootModule().Resources["uptimekuma_notification_wecom.test"]
	return rs.Primary.Attributes["id"], nil
}

func testAccNotificationWeComResourceConfig(name string, botKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wecom" "test" {
  name    = %[1]q
  is_active = true
  bot_key = %[2]q
}
`, name, botKey)
}
