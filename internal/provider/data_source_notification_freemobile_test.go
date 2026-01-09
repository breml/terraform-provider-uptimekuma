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

func TestAccNotificationFreemobileDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationFreemobile")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFreemobileDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_freemobile.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationFreemobileDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_freemobile.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationFreemobileDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_freemobile" "test" {
  name      = %[1]q
  is_active = true
  user      = "1234567890"
  pass      = "test_api_key"
}

data "uptimekuma_notification_freemobile" "test" {
  name = uptimekuma_notification_freemobile.test.name
}
`, name)
}

func testAccNotificationFreemobileDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_freemobile" "test" {
  name      = %[1]q
  is_active = true
  user      = "1234567890"
  pass      = "test_api_key"
}

data "uptimekuma_notification_freemobile" "test" {
  id = uptimekuma_notification_freemobile.test.id
}
`, name)
}
