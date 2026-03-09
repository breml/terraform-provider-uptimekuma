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

func TestAccNotificationOneBotDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationOneBot")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOneBotDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_onebot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationOneBotDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_onebot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationOneBotDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onebot" "test" {
  name        = %[1]q
  is_active   = true
  http_addr   = "http://localhost:5700"
  msg_type    = "group"
  receiver_id = "123456789"
}

data "uptimekuma_notification_onebot" "test" {
  name = uptimekuma_notification_onebot.test.name
}
`, name)
}

func testAccNotificationOneBotDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_onebot" "test" {
  name        = %[1]q
  is_active   = true
  http_addr   = "http://localhost:5700"
  msg_type    = "group"
  receiver_id = "123456789"
}

data "uptimekuma_notification_onebot" "test" {
  id = uptimekuma_notification_onebot.test.id
}
`, name)
}
