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

func TestAccNotificationPushbulletDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationPushbullet")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushbulletDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationPushbulletDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pushbullet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPushbulletDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushbullet" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "o.test1234567890abcdefghijklmnopqrst"
}

data "uptimekuma_notification_pushbullet" "test" {
  name = uptimekuma_notification_pushbullet.test.name
}
`, name)
}

func testAccNotificationPushbulletDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushbullet" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "o.test1234567890abcdefghijklmnopqrst"
}

data "uptimekuma_notification_pushbullet" "test" {
  id = uptimekuma_notification_pushbullet.test.id
}
`, name)
}
