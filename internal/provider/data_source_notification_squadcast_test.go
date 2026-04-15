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

func TestAccNotificationSquadcastDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSquadcast")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSquadcastDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_squadcast.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSquadcastDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_squadcast.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSquadcastDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_squadcast" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://api.squadcast.com/v3/incidents/webhook/test"
}

data "uptimekuma_notification_squadcast" "test" {
  name = uptimekuma_notification_squadcast.test.name
}
`, name)
}

func testAccNotificationSquadcastDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_squadcast" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://api.squadcast.com/v3/incidents/webhook/test"
}

data "uptimekuma_notification_squadcast" "test" {
  id = uptimekuma_notification_squadcast.test.id
}
`, name)
}
