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

func TestAccNotificationSMSPlanetDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSMSPlanet")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSMSPlanetDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSMSPlanetDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_smsplanet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSMSPlanetDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsplanet" "test" {
  name          = %[1]q
  is_active     = true
  api_token     = "test-api-token-datasource"
  phone_numbers = "+48123456789"
}

data "uptimekuma_notification_smsplanet" "test" {
  name = uptimekuma_notification_smsplanet.test.name
}
`, name)
}

func testAccNotificationSMSPlanetDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_smsplanet" "test" {
  name          = %[1]q
  is_active     = true
  api_token     = "test-api-token-datasource"
  phone_numbers = "+48123456789"
}

data "uptimekuma_notification_smsplanet" "test" {
  id = uptimekuma_notification_smsplanet.test.id
}
`, name)
}
