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

func TestAccNotificationPagerDutyDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationPagerDuty")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPagerDutyDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_pagerduty.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationPagerDutyDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_pagerduty" "test" {
  name              = %[1]q
  is_active         = true
  integration_url   = "https://events.pagerduty.com/integration/test/enqueue"
  integration_key   = "test-key-123456789"
  priority          = "high"
  auto_resolve      = "resolved"
}

data "uptimekuma_notification_pagerduty" "test" {
  name = uptimekuma_notification_pagerduty.test.name
}
`, name)
}
