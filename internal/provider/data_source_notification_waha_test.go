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

func TestAccNotificationWAHADataSource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWAHA")
	apiURL := "https://api.waha.local:3000"
	session := "default"
	chatID := "120363101234567890@g.us"
	apiKey := "test-api-key-123"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWAHADataSourceByNameConfig(
					name,
					apiURL,
					session,
					chatID,
					apiKey,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_waha.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			{
				Config: testAccNotificationWAHADataSourceByIDConfig(
					name,
					apiURL,
					session,
					chatID,
					apiKey,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.uptimekuma_notification_waha.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
		},
	})
}

func testAccNotificationWAHADataSourceByNameConfig(
	name string,
	apiURL string,
	session string,
	chatID string,
	apiKey string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_waha" "test" {
  name      = %[1]q
  is_active = true
  api_url   = %[2]q
  session   = %[3]q
  chat_id   = %[4]q
  api_key   = %[5]q
}

data "uptimekuma_notification_waha" "test" {
  name = uptimekuma_notification_waha.test.name
}
`, name, apiURL, session, chatID, apiKey)
}

func testAccNotificationWAHADataSourceByIDConfig(
	name string,
	apiURL string,
	session string,
	chatID string,
	apiKey string,
) string {
	return providerConfig() + fmt.Sprintf(`
resource "uptimekuma_notification_waha" "test" {
  name      = %[1]q
  is_active = true
  api_url   = %[2]q
  session   = %[3]q
  chat_id   = %[4]q
  api_key   = %[5]q
}

data "uptimekuma_notification_waha" "test" {
  id = uptimekuma_notification_waha.test.id
}
`, name, apiURL, session, chatID, apiKey)
}
