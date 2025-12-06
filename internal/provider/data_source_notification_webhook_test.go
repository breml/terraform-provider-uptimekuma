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

func TestAccNotificationWebhookDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationWebhook")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebhookDataSourceConfig(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_notification_webhook.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
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

data "uptimekuma_notification_webhook" "test" {
  name = uptimekuma_notification_webhook.test.name
}
`, name)
}

func TestAccNotificationWebhookDataSourceByID(t *testing.T) {
	name := acctest.RandomWithPrefix("TestNotificationWebhook")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWebhookDataSourceConfigByID(name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.uptimekuma_notification_webhook.test", tfjsonpath.New("name"), knownvalue.StringExact(name)),
				},
			},
		},
	})
}

func testAccNotificationWebhookDataSourceConfigByID(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_webhook" "test" {
  name        = %[1]q
  is_active   = true
  webhook_url = "https://example.com/webhook"
}

data "uptimekuma_notification_webhook" "test" {
  id = uptimekuma_notification_webhook.test.id
}
`, name)
}
