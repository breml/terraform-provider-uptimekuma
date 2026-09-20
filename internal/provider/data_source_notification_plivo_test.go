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

func TestAccNotificationPlivoDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationPlivo")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPlivoDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_plivo.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_plivo.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_plivo.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPlivoDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_plivo" "test" {
  name        = %[1]q
  is_active   = true
  auth_id     = "MAXXXXXXXXXXXXXXXXXX"
  auth_token  = "authtoken0000000000000000000000000000000"
  from_number = "+15550001111"
  to_number   = "+15550002222"
}

data "uptimekuma_notification_plivo" "by_name" {
  name = uptimekuma_notification_plivo.test.name
}

data "uptimekuma_notification_plivo" "by_id" {
  id = uptimekuma_notification_plivo.test.id
}
`, name)
}
