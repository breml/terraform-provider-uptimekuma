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

func TestAccNotificationHeiiOnCallDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestHeiiOnCallDS")
	apiKey := acctest.RandStringFromCharSet(32, acctest.CharSetAlphaNum)
	triggerID := acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationHeiiOnCallDataSourceConfig(name, apiKey, triggerID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_id",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_heiioncall.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationHeiiOnCallDataSourceConfig(
	name string,
	apiKey string,
	triggerID string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_heiioncall" "test" {
  name       = %[1]q
  api_key    = %[2]q
  trigger_id = %[3]q
}

data "uptimekuma_notification_heiioncall" "by_id" {
  id = uptimekuma_notification_heiioncall.test.id
}

data "uptimekuma_notification_heiioncall" "by_name" {
  name = uptimekuma_notification_heiioncall.test.name
}
`, name, apiKey, triggerID)
}
