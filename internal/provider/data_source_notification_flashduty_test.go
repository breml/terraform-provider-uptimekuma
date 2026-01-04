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

func TestAccNotificationFlashDutyDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationFlashDuty")
	integrationKey := "https://api.flashduty.com/webhook/events/12345678901234567890"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationFlashDutyDataSourceConfig(name, integrationKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_flashduty.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_flashduty.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccNotificationFlashDutyDataSourceConfig(name string, integrationKey string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_flashduty" "test" {
  name            = %[1]q
  is_active       = true
  integration_key = %[2]q
  severity        = "Critical"
}

data "uptimekuma_notification_flashduty" "test" {
  name = uptimekuma_notification_flashduty.test.name
}
`, name, integrationKey)
}
