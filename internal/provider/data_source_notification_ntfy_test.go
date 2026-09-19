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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_ntfy.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_ntfy.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
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

data "uptimekuma_notification_ntfy" "by_name" {
  name = uptimekuma_notification_ntfy.test.name
}

data "uptimekuma_notification_ntfy" "by_id" {
  id = uptimekuma_notification_ntfy.test.id
}
`, name)
}
