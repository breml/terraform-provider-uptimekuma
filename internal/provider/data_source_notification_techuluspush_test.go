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

func TestAccNotificationTechulusPushDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationTechulusPush")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTechulusPushDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationTechulusPushDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_techuluspush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationTechulusPushDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_techuluspush" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-datasource"
}

data "uptimekuma_notification_techuluspush" "test" {
  name = uptimekuma_notification_techuluspush.test.name
}
`, name)
}

func testAccNotificationTechulusPushDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_techuluspush" "test" {
  name      = %[1]q
  is_active = true
  api_key   = "test-api-key-datasource"
}

data "uptimekuma_notification_techuluspush" "test" {
  id = uptimekuma_notification_techuluspush.test.id
}
`, name)
}
