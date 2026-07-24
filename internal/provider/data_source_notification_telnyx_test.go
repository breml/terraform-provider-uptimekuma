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

func TestAccNotificationTelnyxDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationTelnyx")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTelnyxDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_telnyx.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_telnyx.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccNotificationTelnyxDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_telnyx.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationTelnyxDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_telnyx" "test" {
  name         = %[1]q
  is_active    = true
  api_key      = "KEY0000000000000000000000000000000"
  phone_number = "+15550001111"
  to_number    = "+15550002222"
}

data "uptimekuma_notification_telnyx" "test" {
  name = uptimekuma_notification_telnyx.test.name
}
`, name)
}

func testAccNotificationTelnyxDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_telnyx" "test" {
  name         = %[1]q
  is_active    = true
  api_key      = "KEY0000000000000000000000000000000"
  phone_number = "+15550001111"
  to_number    = "+15550002222"
}

data "uptimekuma_notification_telnyx" "test" {
  id = uptimekuma_notification_telnyx.test.id
}
`, name)
}
