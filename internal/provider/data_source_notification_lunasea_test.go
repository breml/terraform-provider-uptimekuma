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

func TestAccNotificationLunaseaDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationLunasea")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationLunaseaDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_lunasea.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationLunaseaDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_lunasea.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationLunaseaDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_lunasea" "test" {
  name            = %[1]q
  is_active       = true
  target          = "user"
  lunasea_user_id = "test_user_123"
}

data "uptimekuma_notification_lunasea" "test" {
  name = uptimekuma_notification_lunasea.test.name
}
`, name)
}

func testAccNotificationLunaseaDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_lunasea" "test" {
  name            = %[1]q
  is_active       = true
  target          = "user"
  lunasea_user_id = "test_user_123"
}

data "uptimekuma_notification_lunasea" "test" {
  id = uptimekuma_notification_lunasea.test.id
}
`, name)
}
