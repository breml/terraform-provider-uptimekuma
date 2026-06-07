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

func TestAccNotificationFluxerDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFluxer")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFluxerDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_fluxer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationFluxerDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_fluxer.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationFluxerDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_fluxer" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://fluxer.example.com/webhook/XXXXXXXX"
}

data "uptimekuma_notification_fluxer" "test" {
  name = uptimekuma_notification_fluxer.test.name
}
`, name)
}

func testAccNotificationFluxerDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_fluxer" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://fluxer.example.com/webhook/XXXXXXXX"
}

data "uptimekuma_notification_fluxer" "test" {
  id = uptimekuma_notification_fluxer.test.id
}
`, name)
}
