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

func TestAccNotificationGoogleSheetsDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationGoogleSheets")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationGoogleSheetsDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_googlesheets.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_googlesheets.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationGoogleSheetsDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_googlesheets.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_googlesheets.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationGoogleSheetsDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_googlesheets" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://script.google.com/macros/s/test/exec"
}

data "uptimekuma_notification_googlesheets" "test" {
  name = uptimekuma_notification_googlesheets.test.name
}
`, name)
}

func testAccNotificationGoogleSheetsDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_googlesheets" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://script.google.com/macros/s/test/exec"
}

data "uptimekuma_notification_googlesheets" "test" {
  id = uptimekuma_notification_googlesheets.test.id
}
`, name)
}
