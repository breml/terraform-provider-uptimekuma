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

func TestAccNotificationWPushDataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWPush")
	apiKey := "test-wpush-api-key-123"
	channel := "test-channel"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWPushDataSourceByNameConfig(
					name,
					apiKey,
					channel,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_wpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationWPushDataSourceByIDConfig(
					name,
					apiKey,
					channel,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_wpush.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationWPushDataSourceByNameConfig(
	name string,
	apiKey string,
	channel string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wpush" "test" {
  name      = %[1]q
  is_active = true
  api_key   = %[2]q
  channel   = %[3]q
}

data "uptimekuma_notification_wpush" "test" {
  name = uptimekuma_notification_wpush.test.name
}
`, name, apiKey, channel)
}

func testAccNotificationWPushDataSourceByIDConfig(
	name string,
	apiKey string,
	channel string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_wpush" "test" {
  name      = %[1]q
  is_active = true
  api_key   = %[2]q
  channel   = %[3]q
}

data "uptimekuma_notification_wpush" "test" {
  id = uptimekuma_notification_wpush.test.id
}
`, name, apiKey, channel)
}
