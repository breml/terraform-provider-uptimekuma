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

func TestAccNotificationTeamsDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationTeams")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTeamsDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_teams.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationTeamsDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_teams.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationTeamsDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_teams" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://example.webhook.office.com/webhookb2/example"
}

data "uptimekuma_notification_teams" "test" {
  name = uptimekuma_notification_teams.test.name
}
`, name)
}

func testAccNotificationTeamsDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_teams" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://example.webhook.office.com/webhookb2/example"
}

data "uptimekuma_notification_teams" "test" {
  id = uptimekuma_notification_teams.test.id
}
`, name)
}
