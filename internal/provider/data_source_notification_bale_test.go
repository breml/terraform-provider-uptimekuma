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

func TestAccNotificationBaleDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationBale")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBaleDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bale.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationBaleDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bale.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationBaleDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bale" "test" {
  name      = %[1]q
  is_active = true
  bot_token = "123456:ABCDEF"
  chat_id   = "111222333"
}

data "uptimekuma_notification_bale" "test" {
  name = uptimekuma_notification_bale.test.name
}
`, name)
}

func testAccNotificationBaleDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bale" "test" {
  name      = %[1]q
  is_active = true
  bot_token = "123456:ABCDEF"
  chat_id   = "111222333"
}

data "uptimekuma_notification_bale" "test" {
  id = uptimekuma_notification_bale.test.id
}
`, name)
}
