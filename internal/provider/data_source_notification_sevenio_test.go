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

func TestAccNotificationSevenioDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationSevenio")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSevenioDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_sevenio.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSevenioDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_sevenio.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSevenioDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_sevenio" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-123"
  sender    = "+491234567890"
  to        = "+491111111111"
}

data "uptimekuma_notification_sevenio" "test" {
  name = uptimekuma_notification_sevenio.test.name
}
`, name)
}

func testAccNotificationSevenioDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_sevenio" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-123"
  sender    = "+491234567890"
  to        = "+491111111111"
}

data "uptimekuma_notification_sevenio" "test" {
  id = uptimekuma_notification_sevenio.test.id
}
`, name)
}
