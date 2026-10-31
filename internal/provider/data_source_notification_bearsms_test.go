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

func TestAccNotificationBearSMSDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationBearSMS")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationBearSMSDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bearsms.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bearsms.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_bearsms.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationBearSMSDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_bearsms" "test" {
  name         = %[1]q
  is_active    = true
  username     = "bearsms-user"
  hash_key     = "bearsms-hash-key"
  phone_number = "41790000000"
}

data "uptimekuma_notification_bearsms" "by_name" {
  name = uptimekuma_notification_bearsms.test.name
}

data "uptimekuma_notification_bearsms" "by_id" {
  id = uptimekuma_notification_bearsms.test.id
}
`, name)
}
