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

func TestAccNotificationOctopushDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationOctopushDS")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOctopushDataSourceConfigByName(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_octopush.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationOctopushDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_octopush.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationOctopushDataSourceConfigByName(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_octopush" "test" {
  name         = %[1]q
  is_active    = true
  version      = "2"
  api_key      = "test-api-key"
  login        = "test-login"
  phone_number = "+1234567890"
}

data "uptimekuma_notification_octopush" "by_name" {
  name = uptimekuma_notification_octopush.test.name
}
`, name)
}

func testAccNotificationOctopushDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_octopush" "test" {
  name         = %[1]q
  is_active    = true
  version      = "2"
  api_key      = "test-api-key"
  login        = "test-login"
  phone_number = "+1234567890"
}

data "uptimekuma_notification_octopush" "by_id" {
  id = uptimekuma_notification_octopush.test.id
}
`, name)
}
