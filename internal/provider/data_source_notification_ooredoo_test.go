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

func TestAccNotificationOoredooDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationOoredoo")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationOoredooDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_ooredoo.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_ooredoo.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_ooredoo.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationOoredooDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_ooredoo" "test" {
  name         = %[1]q
  is_active    = true
  username     = "ooredoo-user"
  access_key   = "accesskey0000000000"
  bearer_token = "bearertoken0000000000000000"
  to_number    = "9601234567"
}

data "uptimekuma_notification_ooredoo" "by_name" {
  name = uptimekuma_notification_ooredoo.test.name
}

data "uptimekuma_notification_ooredoo" "by_id" {
  id = uptimekuma_notification_ooredoo.test.id
}
`, name)
}
