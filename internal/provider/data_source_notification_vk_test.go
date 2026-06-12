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

func TestAccNotificationVKDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationVK")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationVKDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_vk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationVKDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_vk.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationVKDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_vk" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "vk1.a.abcdefghijklmnopqrstuvwxyz"
  peer_id      = "12345"
}

data "uptimekuma_notification_vk" "test" {
  name = uptimekuma_notification_vk.test.name
}
`, name)
}

func testAccNotificationVKDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_vk" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "vk1.a.abcdefghijklmnopqrstuvwxyz"
  peer_id      = "12345"
}

data "uptimekuma_notification_vk" "test" {
  id = uptimekuma_notification_vk.test.id
}
`, name)
}
