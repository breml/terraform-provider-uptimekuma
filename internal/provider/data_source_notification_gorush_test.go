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

func TestAccNotificationGorushDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationGorush")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGorushDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_gorush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationGorushDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_gorush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationGorushDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gorush" "test" {
  name         = %[1]q
  is_active    = true
  server_url   = "https://gorush.example.com"
  device_token = "test-device-token"
  platform     = "ios"
}

data "uptimekuma_notification_gorush" "test" {
  name = uptimekuma_notification_gorush.test.name
}
`, name)
}

func testAccNotificationGorushDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_gorush" "test" {
  name         = %[1]q
  is_active    = true
  server_url   = "https://gorush.example.com"
  device_token = "test-device-token"
  platform     = "ios"
}

data "uptimekuma_notification_gorush" "test" {
  id = uptimekuma_notification_gorush.test.id
}
`, name)
}
