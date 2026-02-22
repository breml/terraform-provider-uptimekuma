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

func TestAccNotificationKeepDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationKeep")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationKeepDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_keep.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationKeepDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_keep.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationKeepDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_keep" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://api.keephq.dev/alerts/alert"
  api_key     = "test-api-key"
}

data "uptimekuma_notification_keep" "test" {
  name = uptimekuma_notification_keep.test.name
}
`, name)
}

func testAccNotificationKeepDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_keep" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://api.keephq.dev/alerts/alert"
  api_key     = "test-api-key"
}

data "uptimekuma_notification_keep" "test" {
  id = uptimekuma_notification_keep.test.id
}
`, name)
}
