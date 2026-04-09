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

func TestAccNotificationSIGNL4DataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationSIGNL4")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationSIGNL4DataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_signl4.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationSIGNL4DataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_signl4.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationSIGNL4DataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signl4" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://connect.signl4.com/webhook/example"
}

data "uptimekuma_notification_signl4" "test" {
  name = uptimekuma_notification_signl4.test.name
}
`, name)
}

func testAccNotificationSIGNL4DataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_signl4" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://connect.signl4.com/webhook/example"
}

data "uptimekuma_notification_signl4" "test" {
  id = uptimekuma_notification_signl4.test.id
}
`, name)
}
