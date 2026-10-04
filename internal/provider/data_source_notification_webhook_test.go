package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccNotificationWebhookDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationWebhook")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebhookDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_webhook.by_name",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_webhook.by_id",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationWebhookDataSourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webhook" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://example.com/webhook"
}

data "uptimekuma_notification_webhook" "by_name" {
  name = uptimekuma_notification_webhook.test.name
}

data "uptimekuma_notification_webhook" "by_id" {
  id = uptimekuma_notification_webhook.test.id
}
`, name)
}

// TestAccNotificationWebhookDataSourceNameMismatch verifies that a lookup
// stating both id and name is served while the two agree and rejected once they
// do not, instead of the id silently winning.
func TestAccNotificationWebhookDataSourceNameMismatch(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationWebhookDSNameMismatch")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebhookDataSourceConfigIDAndName(name, name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_webhook.both",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config:      testAccNotificationWebhookDataSourceConfigIDAndName(name, name+"-other"),
				ExpectError: regexp.MustCompile(`Notification name mismatch`),
			},
		},
	})
}

func testAccNotificationWebhookDataSourceConfigIDAndName(name string, lookupName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webhook" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://example.com/webhook"
}

data "uptimekuma_notification_webhook" "both" {
  id   = uptimekuma_notification_webhook.test.id
  name = %[2]q
}
`, name, lookupName)
}
