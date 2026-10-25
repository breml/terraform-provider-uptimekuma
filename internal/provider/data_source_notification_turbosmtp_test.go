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

func TestAccNotificationTurboSMTPDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationTurboSMTP")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTurboSMTPDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_turbosmtp.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_turbosmtp.by_name",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_turbosmtp.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationTurboSMTPDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_turbosmtp" "test" {
  name            = %[1]q
  is_active       = true
  consumer_key    = "consumer-key"
  consumer_secret = "consumer-secret"
  from_email      = "alerts@example.com"
  to_email        = "oncall@example.com"
}

data "uptimekuma_notification_turbosmtp" "by_name" {
  name = uptimekuma_notification_turbosmtp.test.name
}

data "uptimekuma_notification_turbosmtp" "by_id" {
  id = uptimekuma_notification_turbosmtp.test.id
}
`, name)
}
