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

func TestAccNotificationNtfyDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationNtfy")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_notification_ntfy.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
			{
				Config: testAccNotificationNtfyDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_notification_ntfy.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccNotificationNtfyDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name      = %[1]q
  is_active = true

  authentication_method = "none"
  server_url            = "https://ntfy.sh"
  priority              = 5
  topic                 = %[1]q
}

data "uptimekuma_notification_ntfy" "test" {
  name = uptimekuma_notification_ntfy.test.name
}
`, name)
}

func testAccNotificationNtfyDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ntfy" "test" {
  name      = %[1]q
  is_active = true

  authentication_method = "none"
  server_url            = "https://ntfy.sh"
  priority              = 5
  topic                 = %[1]q
}

data "uptimekuma_notification_ntfy" "test" {
  id = uptimekuma_notification_ntfy.test.id
}
`, name)
}
