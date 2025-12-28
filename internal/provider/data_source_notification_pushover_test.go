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

func TestAccNotificationPushoverDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationPushover")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushoverDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pushover.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationPushoverDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pushover.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPushoverDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushover" "test" {
  name      = %[1]q
  is_active = true
  user_key  = "test-user-key-123"
  app_token = "test-app-token-789"
}

data "uptimekuma_notification_pushover" "test" {
  name = uptimekuma_notification_pushover.test.name
}
`, name)
}

func testAccNotificationPushoverDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pushover" "test" {
  name      = %[1]q
  is_active = true
  user_key  = "test-user-key-123"
  app_token = "test-app-token-789"
}

data "uptimekuma_notification_pushover" "test" {
  id = uptimekuma_notification_pushover.test.id
}
`, name)
}
