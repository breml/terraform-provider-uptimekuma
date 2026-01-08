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

func TestAccNotificationLinenotifyDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationLinenotify")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLinenotifyDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_linenotify.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationLinenotifyDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_linenotify.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationLinenotifyDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_linenotify" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "test-access-token-12345"
}

data "uptimekuma_notification_linenotify" "test" {
  name = uptimekuma_notification_linenotify.test.name
}
`, name)
}

func testAccNotificationLinenotifyDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_linenotify" "test" {
  name         = %[1]q
  is_active    = true
  access_token = "test-access-token-12345"
}

data "uptimekuma_notification_linenotify" "test" {
  id = uptimekuma_notification_linenotify.test.id
}
`, name)
}
