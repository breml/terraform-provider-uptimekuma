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

func TestAccNotificationWAHAResource(t *testing.T) {
	name := acctest.RandomWithPrefix("NotificationWAHA")
	nameUpdated := acctest.RandomWithPrefix("NotificationWAHAUpdated")
	apiURL := "https://api.waha.local:3000"
	apiURLUpdated := "https://api-new.waha.local:3000"
	session := "default"
	sessionUpdated := "production"
	chatID := "120363101234567890@g.us"
	chatIDUpdated := "120363109876543210@g.us"
	apiKey := "test-api-key-123"
	apiKeyUpdated := "test-api-key-456"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationWAHAResourceConfig(
					name,
					apiURL,
					session,
					chatID,
					apiKey,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("api_url"),
						knownvalue.StringExact(apiURL),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("session"),
						knownvalue.StringExact(session),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact(chatID),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKey),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccNotificationWAHAResourceConfig(
					nameUpdated,
					apiURLUpdated,
					sessionUpdated,
					chatIDUpdated,
					apiKeyUpdated,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(nameUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("api_url"),
						knownvalue.StringExact(apiURLUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("session"),
						knownvalue.StringExact(sessionUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("chat_id"),
						knownvalue.StringExact(chatIDUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("api_key"),
						knownvalue.StringExact(apiKeyUpdated),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_notification_waha.test",
						tfjsonpath.New("is_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_notification_waha.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

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
				Config: testAccNotificationWAHADataSourceConfig(
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

func testAccNotificationWAHAResourceConfig(
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
`, name, apiURL, session, chatID, apiKey)
}

func testAccNotificationWAHADataSourceConfig(
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
